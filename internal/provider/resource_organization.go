// SPECIAL: This file may require repo or controller-specific things.
package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	urlParser "net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tfbrew/terraform-provider-aap/internal/configprefix"
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
				Computed:    true,
				Description: "Organization ID from the controller API. See `gateway_id` or `eda_id` for the organization's ID in the other APIs.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
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
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the organization.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Organization description.",
			},
			"default_environment": schema.Int32Attribute{
				Optional:    true,
				Description: "AWX/AAP2.4 only. The fallback execution environment that will be used for jobs inside of this organization if not explicitly assigned at the project, job template or workflow level.",
			},
			"max_hosts": schema.Int32Attribute{
				Optional:    true,
				Description: "AWX/AAP2.4 only Maximum number of hosts allowed to be managed by this organization.",
				Default:     int32default.StaticInt32(0),
				Computed:    true,
			},
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
	if !(data.DefaultEnv.IsNull()) {
		bodyData.DefaultEnv = int(data.DefaultEnv.ValueInt32())
	}
	if !(data.MaxHosts.IsNull()) {
		bodyData.MaxHosts = int(data.MaxHosts.ValueInt32())
	}

	url := "organizations/"
	returnedData, _, err := r.client.CreateUpdateAPIRequest(ctx, http.MethodPost, url, bodyData, []int{201}, "gateway")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API http request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}

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

		// overwrite returnedData with Get against org's /controller/ endpoint

		url := fmt.Sprintf("organizations/?name=%s", urlParser.QueryEscape(data.Name.ValueString()))
		responseBodyData, _, err := r.client.GenericAPIRequest(ctx, http.MethodGet, url, nil, []int{200}, "controller")
		if err != nil {
			resp.Diagnostics.AddError(
				"Error making API http request",
				fmt.Sprintf("Error was: %s.", err.Error()))
			return
		}
		// get id
		var nameResult JTChildAPIRead
		err = json.Unmarshal(responseBodyData, &nameResult)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to unmarshal response body into result object",
				fmt.Sprintf("Error:  %v.", err.Error()))
			return
		}
		if nameResult.Count != 1 {
			resp.Diagnostics.AddError(
				"Org controller result count not 1.",
				fmt.Sprintf("Querying for org by name against controller endpoint resulted in result count of %d isntead of 1.", nameResult.Count))
			return
		}
		data.Id = types.StringValue(strconv.Itoa(nameResult.Results[0].Id))

		eda_url := fmt.Sprintf("organizations/?name=%s", urlParser.QueryEscape(data.Name.ValueString()))
		body, edaStatusCode, err := r.client.GenericAPIRequest(ctx, http.MethodGet, eda_url, nil, []int{200, 404}, "eda")
		if err != nil {
			resp.Diagnostics.AddError(
				"Error making API http request",
				fmt.Sprintf("Error was: %s.", err.Error()))
			return
		}

		// If the eda endpoint exists (200), parse the response and extract the EDA ID. If it doesn't exist (404), just continue without error and leave the EDA ID empty.
		if edaStatusCode == 200 {
			// Parse EDA response and extract ID
			var edaResult JTChildAPIRead
			err = json.Unmarshal(body, &edaResult)
			if err != nil {
				resp.Diagnostics.AddError(
					"Unable to unmarshal EDA response body into result object",
					fmt.Sprintf("Error: %v.", err.Error()))
				return
			}
			if edaResult.Count != 1 {
				resp.Diagnostics.AddError(
					"Org EDA result count not 1.",
					fmt.Sprintf("Querying for org by name against EDA endpoint resulted in result count of %d instead of 1.", edaResult.Count))
				return
			}
			data.EdaId = types.Int32Value(int32(edaResult.Results[0].Id))
		} else {
			data.EdaId = types.Int32Null()
		}

	} else {
		data.Id = types.StringValue(fmt.Sprintf("%v", returnedData["id"]))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
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

	if !data.DefaultEnv.IsNull() || responseData.DefaultEnv != 0 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("default_environment"), responseData.DefaultEnv)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("max_hosts"), responseData.MaxHosts)...)

	// if aap2.5+ get the /gateway/ id and set the related field
	if configprefix.Prefix == "aap" {

		gatewayUrl := fmt.Sprintf("organizations/?name=%s", urlParser.QueryEscape(responseData.Name))
		gatewayBody, _, err := r.client.GenericAPIRequest(ctx, http.MethodGet, gatewayUrl, nil, []int{200}, "gateway")
		if err != nil {
			resp.Diagnostics.AddError(
				"Error making API http request",
				fmt.Sprintf("Error was: %s.", err.Error()))
			return
		}

		var nameResult JTChildAPIRead
		err = json.Unmarshal(gatewayBody, &nameResult)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to unmarshal response body into result object",
				fmt.Sprintf("Error:  %v.", err.Error()))
			return
		}
		if nameResult.Count != 1 {
			resp.Diagnostics.AddError(
				"Expected only one org result from gateway",
				fmt.Sprintf("Got count of %d instead.", nameResult.Count))
			return
		}

		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("gateway_id"), nameResult.Results[0].Id)...)

		edaUrl := fmt.Sprintf("organizations/?name=%s", urlParser.QueryEscape(responseData.Name))
		edaBody, edaStatusCode, err := r.client.GenericAPIRequest(ctx, http.MethodGet, edaUrl, nil, []int{200, 404}, "eda")
		if err != nil {
			resp.Diagnostics.AddError(
				"Error making API http request",
				fmt.Sprintf("Error was: %s.", err.Error()))
			return
		}

		// If the eda endpoint exists (200), parse the response and extract the EDA ID. If it doesn't exist (404), just continue without error and leave the EDA ID empty.
		if edaStatusCode == 200 {
			// Parse EDA response and extract ID
			var edaResult JTChildAPIRead
			err = json.Unmarshal(edaBody, &edaResult)
			if err != nil {
				resp.Diagnostics.AddError(
					"Unable to unmarshal EDA response body into result object",
					fmt.Sprintf("Error: %v.", err.Error()))
				return
			}
			if edaResult.Count != 1 {
				resp.Diagnostics.AddError(
					"Org EDA result count not 1.",
					fmt.Sprintf("Querying for org by name against EDA endpoint resulted in result count of %d instead of 1.", edaResult.Count))
				return
			}

			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("eda_id"), edaResult.Results[0].Id)...)
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
}

func (r *OrganizationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data OrganizationModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var id int
	var err error

	if configprefix.Prefix == "aap" {
		id = int(data.GatewayId.ValueInt32())
	} else {
		id, err = strconv.Atoi(data.Id.ValueString())
	}

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
	if !(data.DefaultEnv.IsNull()) {
		bodyData.DefaultEnv = int(data.DefaultEnv.ValueInt32())
	}
	if !(data.MaxHosts.IsNull()) {
		bodyData.MaxHosts = int(data.MaxHosts.ValueInt32())
	}

	url := fmt.Sprintf("organizations/%d/", id)
	_, _, err = r.client.CreateUpdateAPIRequest(ctx, http.MethodPut, url, bodyData, []int{200}, "gateway")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API update request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data OrganizationModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var id int
	var err error

	if configprefix.Prefix == "aap" {
		id = int(data.GatewayId.ValueInt32())
	} else {
		id, err = strconv.Atoi(data.Id.ValueString())
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable convert id from string to int",
			fmt.Sprintf("Unable to convert id: %v.", data.Id.ValueString()))
		return
	}
	url := fmt.Sprintf("organizations/%d/", id)

	_, _, err = r.client.GenericAPIRequest(ctx, http.MethodDelete, url, nil, []int{202, 204}, "gateway")
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
