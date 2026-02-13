package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &EdaDecisionEnvironmentResource{}
var _ resource.ResourceWithImportState = &EdaDecisionEnvironmentResource{}

func NewEdaDecisionEnvironmentResource() resource.Resource {
	return &EdaDecisionEnvironmentResource{}
}

type EdaDecisionEnvironmentResource struct {
	client *providerClient
}

func (r *EdaDecisionEnvironmentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_eda_decision_environment"
}

func (r *EdaDecisionEnvironmentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an EDA Decision Environment resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the EDA Decision Environment.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the EDA Decision Environment.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "The description of the EDA Decision Environment.",
			},
			"eda_credential_id": schema.Int32Attribute{
				Optional:    true,
				Description: "The EDA credential ID associated with the EDA Decision Environment.",
			},
			"organization_id": schema.Int32Attribute{
				Required:    true,
				Description: "The organization ID for the EDA Decision Environment.",
			},
			"image_url": schema.StringAttribute{
				Optional:    true,
				Description: "The image URL for the EDA Decision Environment.",
			},
			"pull_policy": schema.StringAttribute{
				Optional:    true,
				Description: "The pull policy for the EDA Decision Environment. Either `always` or `never`.",
				Validators: []validator.String{
					stringvalidator.OneOf("always", "never"),
				},
			},
		},
	}
}

func (r *EdaDecisionEnvironmentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *EdaDecisionEnvironmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data EdaDecisionEnvironmentModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var bodyData EdaDecisionEnvironmentAPIModel

	bodyData.Name = data.Name.ValueString()
	bodyData.OrganizationId = int(data.OrganizationId.ValueInt32())
	bodyData.ImageUrl = data.ImageUrl.ValueString()
	bodyData.PullPolicy = data.PullPolicy.ValueString()

	if !(data.Description.IsNull()) {
		bodyData.Description = data.Description.ValueString()
	}

	if !(data.EdaCredentialId.IsNull()) {
		bodyData.EdaCredentialId = int(data.EdaCredentialId.ValueInt32())
	}

	url := "decision-environments/"
	returnedData, _, err := r.client.CreateUpdateAPIRequest(ctx, http.MethodPost, url, bodyData, []int{201}, "eda")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API http request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}

	data.Id = types.StringValue(fmt.Sprintf("%v", returnedData["id"]))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EdaDecisionEnvironmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data EdaDecisionEnvironmentModel

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

	url := fmt.Sprintf("decision-environments/%d/", id)
	body, statusCode, err := r.client.GenericAPIRequest(ctx, http.MethodGet, url, nil, []int{200, 404}, "eda")
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

	var responseData EdaDecisionEnvironmentAPIModel

	err = json.Unmarshal(body, &responseData)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to unmarshal json",
			fmt.Sprintf("bodyData: %+v.", body))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), responseData.Name)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), responseData.Organization.Id)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("image_url"), responseData.ImageUrl)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("pull_policy"), responseData.PullPolicy)...)

	if !data.Description.IsNull() || responseData.Description != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("description"), responseData.Description)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if responseData.EdaCredential.Id != 0 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("eda_credential_id"), responseData.EdaCredential.Id)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
}

func (r *EdaDecisionEnvironmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data EdaDecisionEnvironmentModel

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

	var bodyData EdaDecisionEnvironmentAPIModel

	bodyData.Name = data.Name.ValueString()
	bodyData.OrganizationId = int(data.OrganizationId.ValueInt32())
	bodyData.ImageUrl = data.ImageUrl.ValueString()
	bodyData.PullPolicy = data.PullPolicy.ValueString()

	if !(data.Description.IsNull()) {
		bodyData.Description = data.Description.ValueString()
	}

	if !(data.EdaCredentialId.IsNull()) {
		bodyData.EdaCredentialId = int(data.EdaCredentialId.ValueInt32())
	}

	url := fmt.Sprintf("decision-environments/%d/", id)
	_, _, err = r.client.CreateUpdateAPIRequest(ctx, http.MethodPatch, url, bodyData, []int{200}, "eda")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API update request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EdaDecisionEnvironmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data EdaDecisionEnvironmentModel

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

	url := fmt.Sprintf("decision-environments/%d/", id)
	_, _, err = r.client.GenericAPIRequest(ctx, http.MethodDelete, url, nil, []int{202, 204}, "eda")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API delete request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}
}

func (r *EdaDecisionEnvironmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
