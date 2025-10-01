package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ConstructedInventoryResource{}
var _ resource.ResourceWithImportState = &ConstructedInventoryResource{}

func NewConstructedInventoryResource() resource.Resource {
	return &ConstructedInventoryResource{}
}

type ConstructedInventoryResource struct {
	client *providerClient
}

func (r *ConstructedInventoryResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_constructed_inventory"
}

func (r *ConstructedInventoryResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `Manage an Automation Controller contructed inventory.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Inventory ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Inventory name.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Inventory description.",
				Optional:    true,
			},
			"organization": schema.Int32Attribute{
				Description: "Organization ID containing this constructed inventory.",
				Required:    true,
			},
			"variables": schema.StringAttribute{
				Description: "Enter inventory variables using either JSON or YAML syntax.",
				Optional:    true,
			},
			"prevent_instance_group_fallback": schema.BoolAttribute{
				Description: "If enabled, the inventory will prevent adding any organization instance groups to the list of preferred instances groups to run associated job templates on.If this setting is enabled and you provided an empty list, the global instance groups will be applied.",
				Optional:    true,
			},
			"source_vars": schema.StringAttribute{
				Description: "The source_vars for the related auto-created inventory source, special to constructed inventory.",
				Optional:    true,
			},
			"update_cache_timeout": schema.Int32Attribute{
				Description: "The cache timeout for the related auto-created inventory source, special to constructed inventory.",
				Optional:    true,
			},
			"limit": schema.StringAttribute{
				Description: "Hosts in the inventory will be limited to only those that match the filter.",
				Optional:    true,
			},
			"verbosity": schema.Int32Attribute{
				Description: "The verbosity level for the related auto-created inventory source, special to constructed inventory",
				Optional:    true,
				Default:     int32default.StaticInt32(0),
				Computed:    true,
				Validators: []validator.Int32{
					int32validator.Between(0, 2),
				},
			},
		},
	}
}

func (r *ConstructedInventoryResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ConstructedInventoryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ConstructedInventoryModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var bodyData ConstructedInventoryAPIModel

	if !(data.Name.IsNull()) {
		bodyData.Name = data.Name.ValueString()
	}
	if !(data.Description.IsNull()) {
		bodyData.Description = data.Description.ValueString()
	}
	if !(data.Organization.IsNull()) {
		bodyData.Organization = int(data.Organization.ValueInt32())
	}
	if !(data.Variables.IsNull()) {
		bodyData.Variables = data.Variables.ValueString()
	}
	if !(data.PreventInstanceGroupFallback.IsNull()) {
		bodyData.PreventInstanceGroupFallback = data.PreventInstanceGroupFallback.ValueBool()
	}
	if !(data.SourceVars.IsNull()) {
		bodyData.SourceVars = data.SourceVars.ValueString()
	}
	if !(data.UpdateCacheTimeout.IsNull()) {
		bodyData.UpdateCacheTimeout = int(data.UpdateCacheTimeout.ValueInt32())
	}
	if !(data.Limit.IsNull()) {
		bodyData.Limit = data.Limit.ValueString()
	}
	if !(data.Verbosity.IsNull()) {
		bodyData.Verbosity = int(data.Verbosity.ValueInt32())
	}

	url := "constructed_inventories/"
	returnedData, _, err := r.client.CreateUpdateAPIRequest(ctx, http.MethodPost, url, bodyData, []int{201}, "")
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

func (r *ConstructedInventoryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ConstructedInventoryModel

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

	url := fmt.Sprintf("constructed_inventories/%d/", id)
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

	var responseData ConstructedInventoryAPIModel

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

	if !data.Organization.IsNull() || responseData.Organization != 0 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization"), responseData.Organization)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !data.Variables.IsNull() || responseData.Variables != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("variables"), responseData.Variables)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !data.PreventInstanceGroupFallback.IsNull() || responseData.PreventInstanceGroupFallback {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("prevent_instance_group_fallback"), responseData.PreventInstanceGroupFallback)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !data.SourceVars.IsNull() || responseData.SourceVars != "" {

		if !strings.HasSuffix(responseData.SourceVars, "\n") {
			responseData.SourceVars += "\n"
		}

		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("source_vars"), responseData.SourceVars)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !data.UpdateCacheTimeout.IsNull() || responseData.UpdateCacheTimeout != 0 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("update_cache_timeout"), responseData.UpdateCacheTimeout)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !data.Limit.IsNull() || responseData.Limit != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("limit"), responseData.Limit)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !data.Verbosity.IsNull() || responseData.Verbosity != 0 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("verbosity"), responseData.Verbosity)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

}

func (r *ConstructedInventoryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ConstructedInventoryModel

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

	var bodyData ConstructedInventoryAPIModel

	if !(data.Name.IsNull()) {
		bodyData.Name = data.Name.ValueString()
	}
	if !(data.Description.IsNull()) {
		bodyData.Description = data.Description.ValueString()
	}
	if !(data.Organization.IsNull()) {
		bodyData.Organization = int(data.Organization.ValueInt32())
	}
	if !(data.Variables.IsNull()) {
		bodyData.Variables = data.Variables.ValueString()
	}
	if !(data.PreventInstanceGroupFallback.IsNull()) {
		bodyData.PreventInstanceGroupFallback = data.PreventInstanceGroupFallback.ValueBool()
	}
	if !(data.SourceVars.IsNull()) {
		bodyData.SourceVars = data.SourceVars.ValueString()
	}
	if !(data.UpdateCacheTimeout.IsNull()) {
		bodyData.UpdateCacheTimeout = int(data.UpdateCacheTimeout.ValueInt32())
	}
	if !(data.Limit.IsNull()) {
		bodyData.Limit = data.Limit.ValueString()
	}
	if !(data.Verbosity.IsNull()) {
		bodyData.Verbosity = int(data.Verbosity.ValueInt32())
	}

	url := fmt.Sprintf("constructed_inventories/%d/", id)
	_, _, err = r.client.CreateUpdateAPIRequest(ctx, http.MethodPut, url, bodyData, []int{200}, "")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API update request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ConstructedInventoryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ConstructedInventoryModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable convert id from string to int",
			fmt.Sprintf("Unable to convert id: %v.", data.Id.ValueString()))
		return
	}

	url := fmt.Sprintf("constructed_inventories/%d/", id)
	_, _, err = r.client.GenericAPIRequest(ctx, http.MethodDelete, url, nil, []int{202, 204}, "")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API delete request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}
}

func (r *ConstructedInventoryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
