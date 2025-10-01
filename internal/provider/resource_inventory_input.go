package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

var _ resource.Resource = &InventoryInputResource{}

var _ resource.ResourceWithImportState = &InventoryInputResource{}

func NewInventoryInputResource() resource.Resource {
	return &InventoryInputResource{}
}

type InventoryInputResource struct {
	client *providerClient
}

func (r *InventoryInputResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_inventory_input"
}

func (r *InventoryInputResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `Associate an existing inventory to a constructed-type inventory.`,
		Attributes: map[string]schema.Attribute{
			"constructed_inventory_id": schema.StringAttribute{
				Description: "The ID of the constructed Inventory to which you want to associate the input_inventory_id.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"input_inventory_id": schema.StringAttribute{
				Description: "The ID of the existing Inventory to associate as an input to the constructed inventory.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *InventoryInputResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *InventoryInputResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data InventoryInputModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var bodyData InventoryInputAssocAPIModel

	inputInvId, err := strconv.Atoi(data.InputInventoryID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("unable to convert input inventory id to int.", fmt.Sprintf("Unable to convert %v to int", data.InputInventoryID.ValueString()))
		return
	}

	bodyData.Id = inputInvId

	constructedInvId, err := strconv.Atoi(data.ConstructedInventoryID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("unable to convert constructed inventory id to int.", fmt.Sprintf("Unable to convert %v to int", data.ConstructedInventoryID.ValueString()))
		return
	}

	url := fmt.Sprintf("inventories/%d/input_inventories/", constructedInvId)
	_, _, err = r.client.GenericAPIRequest(ctx, http.MethodPost, url, bodyData, []int{204}, "")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API http request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}

func (r *InventoryInputResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data InventoryInputModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	inputInvId, err := strconv.Atoi(data.InputInventoryID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("unable to convert input inventory id to int in read.", fmt.Sprintf("Unable to convert %v to int", data.InputInventoryID.ValueString()))
		return
	}

	constructedInvId, err := strconv.Atoi(data.ConstructedInventoryID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("unable to convert constructed inventory id to int in read.", fmt.Sprintf("Unable to convert %v to int", data.ConstructedInventoryID.ValueString()))
		return
	}

	url := fmt.Sprintf("inventories/%d/input_inventories/?id=%d", constructedInvId, inputInvId)
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

	result := struct {
		Count int `json:"count"`
	}{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to unmarshal response body into object",
			fmt.Sprintf("Error:  %v.", err.Error()))
		return
	}

	if result.Count == 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	if result.Count != 1 {

		resp.Diagnostics.AddError(
			"Incorrect number of input_inventories returned by Ids",
			fmt.Sprintf("Unable to read inventories/id/input_inventories/?id=x as API returned %v results.", result.Count))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InventoryInputResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data InventoryInputModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// NOTE
	// because we have this resource scheme setup to require replace, the update method is intentially bare-minimum

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InventoryInputResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var bodyData ChildDissasocBody
	var data InventoryInputModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	inputInvId, err := strconv.Atoi(data.InputInventoryID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("unable to convert input inventory id to int in Delete.", fmt.Sprintf("Unable to convert %v to int", data.InputInventoryID.ValueString()))
		return
	}

	constructedInvId, err := strconv.Atoi(data.ConstructedInventoryID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("unable to convert constructed inventory id to int in Delete.", fmt.Sprintf("Unable to convert %v to int", data.ConstructedInventoryID.ValueString()))
		return
	}

	url := fmt.Sprintf("inventories/%d/input_inventories/", constructedInvId)

	bodyData.Id = inputInvId
	bodyData.Disassociate = true

	_, _, err = r.client.GenericAPIRequest(ctx, http.MethodPost, url, bodyData, []int{204}, "")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API delete request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}
}

func (r *InventoryInputResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	idUnescaped, _ := strconv.Unquote(`"` + req.ID + `"`)

	idParts := strings.Split(idUnescaped, ",")
	countParts := len(idParts)

	switch countParts {
	case 2:
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("constructed_inventory_id"), idParts[0])...)
		if resp.Diagnostics.HasError() {
			return
		}

		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("input_inventory_id"), idParts[1])...)
		if resp.Diagnostics.HasError() {
			return
		}
	default:
		resp.Diagnostics.AddError("Invalid import id string", "The import string at the end must contain two integers separated by a comma with no spaces. The first is interpreted as the constructed_inventory_id and the second is interpreted as the input_inventory_id. For example: \"23,45\". You provided: "+req.ID)

	}
}
