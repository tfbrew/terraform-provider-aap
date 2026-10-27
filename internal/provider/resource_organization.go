package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	// AAP-AWX-DIFF-BEGIN: aap-only imports for the gateway and EDA IDs and the controller-only fields.
	"github.com/hashicorp/terraform-plugin-framework/diag"
	// AAP-AWX-DIFF-END
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	// AAP-AWX-DIFF-BEGIN: aap-only imports for the gateway and EDA IDs and the controller-only fields.
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	// AAP-AWX-DIFF-END
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	// AAP-AWX-DIFF-BEGIN: aap-only imports for the gateway and EDA IDs and the controller-only fields.
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	// AAP-AWX-DIFF-END
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	// AAP-AWX-DIFF-BEGIN: aap-only imports for the gateway and EDA IDs and the controller-only fields.
	"github.com/tfbrew/terraform-provider-aap/internal/configprefix"
	// AAP-AWX-DIFF-END
)

var _ resource.Resource = &OrganizationResource{}
var _ resource.ResourceWithImportState = &OrganizationResource{}

func NewOrganizationResource() resource.Resource {
	return &OrganizationResource{}
}

type OrganizationResource struct {
	client *providerClient
}

func (r *OrganizationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization"
}

func (r *OrganizationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `Manage an Automation Controller organization.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				// AAP-AWX-DIFF-BEGIN: aap documents that id is the controller ID.
				Computed:    true,
				Description: "Organization ID from the controller API. See `gateway_id` or `eda_id` for the organization's ID in the other APIs.",
				// AAP-AWX-DIFF-END
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			// AAP-AWX-DIFF-BEGIN: aap also exposes the gateway and EDA organization IDs.
			"gateway_id": schema.Int32Attribute{
				Description: "Organization ID in the gateway API",
				Computed:    true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
			},
			"eda_id": schema.Int32Attribute{
				Description: "Organization ID in the EDA API",
				Computed:    true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
			},
			// AAP-AWX-DIFF-END
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the organization.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Organization description.",
			},
			"max_hosts": schema.Int32Attribute{
				Optional:    true,
				Description: "Maximum number of hosts allowed to be managed by this organization.",
				Default:     int32default.StaticInt32(0),
				Computed:    true,
			},
			// AAP-AWX-DIFF-BEGIN: aap manages opa_query_path; awx manages default_environment (AAP 2.5+ dropped the organization default execution environment).
			"opa_query_path": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "The query path of the OPA policy evaluated before jobs in this organization run, formatted as `package/rule`. Requires an automation controller with policy enforcement support.",
			},
			// AAP-AWX-DIFF-END
		},
	}
}

func (r *OrganizationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	configureData, ok := req.ProviderData.(*providerClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = configureData
}

func (r *OrganizationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data OrganizationModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var bodyData OrganizationAPIModel

	if !(data.Name.IsNull()) {
		bodyData.Name = data.Name.ValueString()
	}
	if !(data.Description.IsNull()) {
		bodyData.Description = data.Description.ValueString()
	}
	if !(data.MaxHosts.IsNull()) {
		bodyData.MaxHosts = int(data.MaxHosts.ValueInt32())
	}
	// AAP-AWX-DIFF-BEGIN: awx also sends default_environment.
	// AAP-AWX-DIFF-END

	url := "organizations/"
	// AAP-AWX-DIFF-BEGIN: aap creates organizations through the gateway API; awx uses /api/v2/.
	returnedData, _, err := r.client.CreateUpdateAPIRequest(ctx, http.MethodPost, url, bodyData, []int{201}, "gateway")
	// AAP-AWX-DIFF-END
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API http request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}

	// AAP-AWX-DIFF-BEGIN: aap looks up the controller and EDA IDs of the new gateway organization (with retries); awx uses the returned ID.
	// organizations must be created using the /gateway/ instead of /controller/ api endpoint. But,
	//  the same org may get 2 different IDs and we need the ID from the controller in order to use
	//  the organization ID.
	id, ok := returnedData["id"].(float64)
	if !ok {
		resp.Diagnostics.AddError(
			"unable to cast ID as float64",
			fmt.Sprintf("Value provided was: %v.", returnedData["id"]))
		return
	}
	data.GatewayId = types.Int32Value(int32(id))

	// now get the EDA ID by querying the eda endpoint

	if configprefix.Prefix == "aap" {

		// the org was just created in gateway, so wait for it to be synced to the controller and EDA
		controllerId, found, err := lookupOrgIdByName(ctx, r.client, data.Name.ValueString(), "controller", true)
		if err == nil && !found {
			err = fmt.Errorf("controller endpoint returned 404")
		}
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to look up organization in controller API",
				fmt.Sprintf("Error was: %s.", err.Error()))
			return
		}
		data.Id = types.StringValue(strconv.Itoa(controllerId))

		// If the eda endpoint doesn't exist (404), just continue without error and leave the EDA ID empty.
		edaId, found, err := lookupOrgIdByName(ctx, r.client, data.Name.ValueString(), "eda", true)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to look up organization in EDA API",
				fmt.Sprintf("Error was: %s.", err.Error()))
			return
		}
		if found {
			data.EdaId = types.Int32Value(int32(edaId))
		} else {
			data.EdaId = types.Int32Null()
		}

	} else {
		data.Id = types.StringValue(fmt.Sprintf("%v", returnedData["id"]))
	}
	// AAP-AWX-DIFF-END

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	// AAP-AWX-DIFF-BEGIN: aap sets the controller-only fields (max_hosts, opa_query_path) through the controller API, since the gateway API ignores them.
	if resp.Diagnostics.HasError() {
		return
	}

	// the gateway API ignores the controller-only fields, so set them through the controller API.
	controllerFields := map[string]any{}
	if data.MaxHosts.ValueInt32() != 0 {
		controllerFields["max_hosts"] = data.MaxHosts.ValueInt32()
	}
	if data.OpaQueryPath.ValueString() != "" {
		controllerFields["opa_query_path"] = data.OpaQueryPath.ValueString()
	}
	if len(controllerFields) > 0 {
		resp.Diagnostics.Append(r.updateControllerFields(ctx, data.Id.ValueString(), controllerFields)...)
	}
	// AAP-AWX-DIFF-END
}

func (r *OrganizationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data OrganizationModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable convert id from string to int",
			fmt.Sprintf("Unable to convert id: %v.", data.Id))
		return
	}

	url := fmt.Sprintf("organizations/%d/", id)
	body, statusCode, err := r.client.GenericAPIRequest(ctx, http.MethodGet, url, nil, []int{200, 404}, "")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API http request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}

	if statusCode == 404 {
		resp.State.RemoveResource(ctx)
		return
	}

	var responseData OrganizationAPIModel

	err = json.Unmarshal(body, &responseData)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to unmarshal json",
			fmt.Sprintf("bodyData: %+v.", body))
		return
	}

	if !data.Name.IsNull() || responseData.Name != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), responseData.Name)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !data.Description.IsNull() || responseData.Description != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("description"), responseData.Description)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("max_hosts"), responseData.MaxHosts)...)
	// AAP-AWX-DIFF-BEGIN: aap reads opa_query_path and looks up the gateway and EDA organization IDs; awx reads default_environment.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("opa_query_path"), responseData.OpaQueryPath)...)

	// if aap2.5+ get the /gateway/ id and set the related field
	if configprefix.Prefix == "aap" {

		gatewayId, found, err := lookupOrgIdByName(ctx, r.client, responseData.Name, "gateway", false)
		if err == nil && !found {
			err = fmt.Errorf("gateway endpoint returned 404")
		}
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to look up organization in gateway API",
				fmt.Sprintf("Error was: %s.", err.Error()))
			return
		}

		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("gateway_id"), gatewayId)...)

		// If the eda endpoint doesn't exist (404), just continue without error and leave the EDA ID empty.
		edaId, found, err := lookupOrgIdByName(ctx, r.client, responseData.Name, "eda", false)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to look up organization in EDA API",
				fmt.Sprintf("Error was: %s.", err.Error()))
			return
		}

		if found {
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("eda_id"), edaId)...)
		} else {
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("eda_id"), types.Int32Null())...)
		}

		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		var id int
		var err error

		id, err = strconv.Atoi(data.Id.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("can't convert Id to int", "unable to convert ID to int.")
			return
		}

		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("gateway_id"), id)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	// AAP-AWX-DIFF-END
}

func (r *OrganizationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data OrganizationModel

	// AAP-AWX-DIFF-BEGIN: aap reads the prior state to find changed controller-only fields.
	var state OrganizationModel

	// AAP-AWX-DIFF-END
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	// AAP-AWX-DIFF-BEGIN: aap reads the prior state to find changed controller-only fields.
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	// AAP-AWX-DIFF-END
	if resp.Diagnostics.HasError() {
		return
	}

	// AAP-AWX-DIFF-BEGIN: aap updates and deletes organizations by their gateway ID; awx uses the /api/v2/ ID.
	var id int
	var err error

	if configprefix.Prefix == "aap" {
		id = int(data.GatewayId.ValueInt32())
	} else {
		id, err = strconv.Atoi(data.Id.ValueString())
	}

	// AAP-AWX-DIFF-END
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable convert id from string to int",
			fmt.Sprintf("Unable to convert id: %v.", data.Id))
		return
	}

	var bodyData OrganizationAPIModel

	if !(data.Name.IsNull()) {
		bodyData.Name = data.Name.ValueString()
	}
	if !(data.Description.IsNull()) {
		bodyData.Description = data.Description.ValueString()
	}
	if !(data.MaxHosts.IsNull()) {
		bodyData.MaxHosts = int(data.MaxHosts.ValueInt32())
	}
	// AAP-AWX-DIFF-BEGIN: awx also sends default_environment.
	// AAP-AWX-DIFF-END

	url := fmt.Sprintf("organizations/%d/", id)
	// AAP-AWX-DIFF-BEGIN: aap updates organizations through the gateway API; awx uses /api/v2/.
	_, _, err = r.client.CreateUpdateAPIRequest(ctx, http.MethodPut, url, bodyData, []int{200}, "gateway")
	// AAP-AWX-DIFF-END
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API update request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}

	// AAP-AWX-DIFF-BEGIN: aap sets the controller-only fields (max_hosts, opa_query_path) through the controller API, since the gateway API ignores them.
	// the gateway API ignores the controller-only fields, so set them through the controller API.
	controllerFields := map[string]any{}
	if !data.MaxHosts.Equal(state.MaxHosts) {
		controllerFields["max_hosts"] = data.MaxHosts.ValueInt32()
	}
	if !data.OpaQueryPath.Equal(state.OpaQueryPath) {
		controllerFields["opa_query_path"] = data.OpaQueryPath.ValueString()
	}
	if len(controllerFields) > 0 {
		resp.Diagnostics.Append(r.updateControllerFields(ctx, data.Id.ValueString(), controllerFields)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// AAP-AWX-DIFF-END
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	// AAP-AWX-DIFF-BEGIN: aap sets the controller-only fields (max_hosts, opa_query_path) through the controller API, since the gateway API ignores them.
}

// updateControllerFields patches fields that only the controller API manages (max_hosts, opa_query_path).
func (r *OrganizationResource) updateControllerFields(ctx context.Context, id string, fields map[string]any) diag.Diagnostics {
	var diags diag.Diagnostics

	url := fmt.Sprintf("organizations/%s/", id)
	returnedData, _, err := r.client.CreateUpdateAPIRequest(ctx, http.MethodPatch, url, fields, []int{200}, "controller")
	if err != nil {
		diags.AddError(
			"Error making API update request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return diags
	}

	if opaQueryPath, _ := fields["opa_query_path"].(string); opaQueryPath != "" {
		if _, ok := returnedData["opa_query_path"]; !ok {
			diags.AddAttributeError(
				path.Root("opa_query_path"),
				"Unsupported attribute",
				"The automation controller did not return opa_query_path; it does not support policy enforcement.")
		}
	}

	return diags
	// AAP-AWX-DIFF-END
}

func (r *OrganizationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data OrganizationModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// AAP-AWX-DIFF-BEGIN: aap updates and deletes organizations by their gateway ID; awx uses the /api/v2/ ID.
	var id int
	var err error

	if configprefix.Prefix == "aap" {
		id = int(data.GatewayId.ValueInt32())
	} else {
		id, err = strconv.Atoi(data.Id.ValueString())
	}

	// AAP-AWX-DIFF-END
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable convert id from string to int",
			fmt.Sprintf("Unable to convert id: %v.", data.Id.ValueString()))
		return
	}
	url := fmt.Sprintf("organizations/%d/", id)

	// AAP-AWX-DIFF-BEGIN: aap deletes organizations through the gateway API; awx uses /api/v2/.
	_, _, err = r.client.GenericAPIRequest(ctx, http.MethodDelete, url, nil, []int{202, 204}, "gateway")
	// AAP-AWX-DIFF-END
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API delete request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}
}

func (r *OrganizationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
