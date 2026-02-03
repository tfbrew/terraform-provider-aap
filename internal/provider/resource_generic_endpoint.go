package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &GenericEndpointResource{}
var _ resource.ResourceWithImportState = &GenericEndpointResource{}

func NewGenericEndpointResource() resource.Resource {
	return &GenericEndpointResource{}
}

type GenericEndpointResource struct {
	client *providerClient
}

func (r *GenericEndpointResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_generic_endpoint"
}

func (r *GenericEndpointResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `WARNING: This resource is still in development and not tested for all AAP endpoints. It can be used as a catch-all for endpoints that do not have a dedicated resource yet. Use at your own risk.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GenericEndpoint ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"api_path": schema.StringAttribute{
				Description: "API path for the GenericEndpoint. If the URL is /api/gateway/v1/authenticators/, the api_path is `authenticators`.",
				Required:    true,
			},
			"api_endpoint": schema.StringAttribute{
				Description: "API endpoint for the GenericEndpoint. If the URL is /api/gateway/v1/authenticators/, the api_endpoint is gateway. Must be one of `eda`, `galaxy`, `gateway`, `controller`.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"eda", "galaxy", "gateway", "controller"}...),
				},
			},
			"data_json": schema.StringAttribute{
				Description: "Data in `jsonencode()` format to send to the API endpoint.",
				Optional:    true,
			},
			"disable_read": schema.BoolAttribute{
				Description: "Set this to true if the api endpoint is not able to read the data or if the endpoint does not support the GET method. Setting this to true will break drift detection. You should set this to true for endpoints that are write-only. Defaults to false.",
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Computed:    true,
			},
			"ignore_read": schema.SetAttribute{
				Description: "List of attribute names to ignore during the read operation. This can be used to prevent drift detection on attributes that are dynamically created with the resource.",
				Optional:    true,
				ElementType: types.StringType,
				Computed:    true,
				Default: setdefault.StaticValue(
					types.SetValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("created"),
							types.StringValue("created_by"),
							types.StringValue("id"),
							types.StringValue("modified"),
							types.StringValue("modified_by"),
							types.StringValue("related"),
							types.StringValue("slug"),
							types.StringValue("summary_fields"),
							types.StringValue("url"),
						},
					),
				),
			},
			"disable_delete": schema.BoolAttribute{
				Description: "Set this to true if the api endpoint is not able to delete the data or if the endpoint does not support the DELETE method. Setting this to true will prevent Terraform from attempting to delete the resource.",
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Computed:    true,
			},
			"update_method": schema.StringAttribute{
				Description: "HTTP method to use for updates. Defaults to `put`. Can be one of `put` or `patch`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("put"),
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"put", "patch"}...),
				},
			},
		},
	}
}

func (r *GenericEndpointResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *GenericEndpointResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data GenericEndpointModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bodyData := data.DataJson.ValueString()
	var jsonData any
	if err := json.Unmarshal([]byte(bodyData), &jsonData); err != nil {
		resp.Diagnostics.AddError(
			"Error parsing JSON data",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}

	url := fmt.Sprintf("%s/", data.ApiPath)
	url = strings.ReplaceAll(url, "\"", "")
	returnedData, _, err := r.client.CreateUpdateAPIRequest(ctx, http.MethodPost, url, jsonData, []int{201}, data.ApiEndpoint.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API http request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}

	returnedValues := []string{"id"}
	for _, key := range returnedValues {
		if _, exists := returnedData[key]; !exists {
			resp.Diagnostics.AddError(
				"Error retrieving computed values",
				fmt.Sprintf("Could not retrieve %v.", key))
			return
		}
	}

	data.Id = types.StringValue(fmt.Sprintf("%v", returnedData["id"]))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GenericEndpointResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data GenericEndpointModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.DisableRead.ValueBool() {
		return
	}

	id, err := strconv.Atoi(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable convert id from string to int",
			fmt.Sprintf("Unable to convert id: %v.", data.Id))
		return
	}

	url := fmt.Sprintf("%s/%d/", data.ApiPath, id)
	url = strings.ReplaceAll(url, "\"", "")
	body, statusCode, err := r.client.GenericAPIRequest(ctx, http.MethodGet, url, nil, []int{200, 404}, data.ApiEndpoint.ValueString())
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

	var bodyMap map[string]any
	if err := json.Unmarshal(body, &bodyMap); err != nil {
		resp.Diagnostics.AddError(
			"Error parsing API response JSON",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}

	var ignoreReadKeys []string
	resp.Diagnostics.Append(data.IgnoreRead.ElementsAs(ctx, &ignoreReadKeys, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, key := range ignoreReadKeys {
		delete(bodyMap, key)
	}

	filteredBody, err := json.Marshal(bodyMap)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error marshaling filtered JSON",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("data_json"), string(filteredBody))...)

}

func (r *GenericEndpointResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data GenericEndpointModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable convert id from string to int",
			fmt.Sprintf("Unable to convert id: %v.", data))
		return
	}

	bodyData := data.DataJson.ValueString()
	var jsonData any
	if err := json.Unmarshal([]byte(bodyData), &jsonData); err != nil {
		resp.Diagnostics.AddError(
			"Error parsing JSON data",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}

	url := fmt.Sprintf("%s/%d/", data.ApiPath, id)
	url = strings.ReplaceAll(url, "\"", "")
	if data.UpdateMethod.ValueString() == "patch" {
		_, _, err = r.client.CreateUpdateAPIRequest(ctx, http.MethodPatch, url, jsonData, []int{200}, data.ApiEndpoint.ValueString())
	} else if data.UpdateMethod.ValueString() == "put" {
		_, _, err = r.client.CreateUpdateAPIRequest(ctx, http.MethodPut, url, jsonData, []int{200}, data.ApiEndpoint.ValueString())
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API update request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GenericEndpointResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data GenericEndpointModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// return early if delete is disabled
	if data.DisableDelete.ValueBool() {
		return
	}

	id, err := strconv.Atoi(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable convert id from string to int",
			fmt.Sprintf("Unable to convert id: %v.", data.Id.ValueString()))
		return
	}

	url := fmt.Sprintf("%s/%d/", data.ApiPath, id)
	url = strings.ReplaceAll(url, "\"", "")
	_, _, err = r.client.GenericAPIRequest(ctx, http.MethodDelete, url, nil, []int{202, 204}, data.ApiEndpoint.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API delete request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}
}

func (r *GenericEndpointResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
