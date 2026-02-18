package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &EdaEventStreamResource{}
var _ resource.ResourceWithImportState = &EdaEventStreamResource{}

func NewEdaEventStreamResource() resource.Resource {
	return &EdaEventStreamResource{}
}

type EdaEventStreamResource struct {
	client *providerClient
}

func (r *EdaEventStreamResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_eda_event_stream"
}

func (r *EdaEventStreamResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an EDA Event Stream resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the EDA Event Stream.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the EDA Event Stream.",
			},
			"additional_data_headers": schema.StringAttribute{
				Optional:    true,
				Description: "The additional http headers which will be added to the event data. The headers are comma delimited.",
			},
			"eda_credential_id": schema.Int32Attribute{
				Required:    true,
				Description: "The EDA credential ID associated with the EDA Event Stream.",
			},
			"organization_id": schema.Int32Attribute{
				Required:    true,
				Description: "The organization ID for the EDA Event Stream.",
			},
			"test_mode": schema.BoolAttribute{
				Optional:    true,
				Description: "Enable test mode.",
				Default:     booldefault.StaticBool(false),
				Computed:    true,
			},
			"uuid": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The uuid for the EDA Event Stream.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"url": schema.StringAttribute{
				Computed:    true,
				Description: "The url for the EDA Event Stream.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"event_stream_type": schema.StringAttribute{
				Computed:    true,
				Description: "The event stream type for the EDA Event Stream based on credential type.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *EdaEventStreamResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data EdaEventStreamModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if data.Uuid.IsNull() {
		return
	} else {
		// Check for valid UUID
		if err := uuid.Validate(data.Uuid.ValueString()); err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("uuid"),
				"Invalid Attribute Configuration",
				"Attribute uuid is invalid. Please provide a valid UUID in the format xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
			)
			return
		}
	}
}

func (r *EdaEventStreamResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *EdaEventStreamResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data EdaEventStreamModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var bodyData EdaEventStreamAPIModel

	bodyData.Name = data.Name.ValueString()
	bodyData.OrganizationId = int(data.OrganizationId.ValueInt32())
	bodyData.EdaCredentialId = int(data.EdaCredentialId.ValueInt32())
	bodyData.TestMode = data.TestMode.ValueBool()

	if !(data.AdditionalDataHeaders.IsNull()) {
		bodyData.AdditionalDataHeaders = data.AdditionalDataHeaders.ValueString()
	}

	if !(data.Uuid.IsNull()) {
		bodyData.Uuid = data.Uuid.ValueString()
	}

	url := "event-streams/"
	returnedData, _, err := r.client.CreateUpdateAPIRequest(ctx, http.MethodPost, url, bodyData, []int{201}, "eda")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API http request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}

	data.Id = types.StringValue(fmt.Sprintf("%v", returnedData["id"]))
	data.Uuid = types.StringValue(fmt.Sprintf("%v", returnedData["uuid"]))
	data.Url = types.StringValue(fmt.Sprintf("%v", returnedData["url"]))
	data.EventStreamType = types.StringValue(fmt.Sprintf("%v", returnedData["event_stream_type"]))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EdaEventStreamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data EdaEventStreamModel

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

	url := fmt.Sprintf("event-streams/%d/", id)
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

	var responseData EdaEventStreamAPIModel

	err = json.Unmarshal(body, &responseData)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to unmarshal json",
			fmt.Sprintf("bodyData: %+v.", body))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), responseData.Name)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), responseData.Organization.Id)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("eda_credential_id"), responseData.EdaCredential.Id)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("test_mode"), responseData.TestMode)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("uuid"), responseData.Uuid)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("url"), responseData.Url)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("event_stream_type"), responseData.EventStreamType)...)

	if !data.AdditionalDataHeaders.IsNull() || responseData.AdditionalDataHeaders != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("additional_data_headers"), responseData.AdditionalDataHeaders)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
}

func (r *EdaEventStreamResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data EdaEventStreamModel

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

	var bodyData EdaEventStreamAPIModel

	bodyData.Name = data.Name.ValueString()
	bodyData.OrganizationId = int(data.OrganizationId.ValueInt32())
	bodyData.EdaCredentialId = int(data.EdaCredentialId.ValueInt32())
	bodyData.TestMode = data.TestMode.ValueBool()

	if !(data.AdditionalDataHeaders.IsNull()) {
		bodyData.AdditionalDataHeaders = data.AdditionalDataHeaders.ValueString()
	}

	if !(data.Uuid.IsNull()) {
		bodyData.Uuid = data.Uuid.ValueString()
	}

	url := fmt.Sprintf("event-streams/%d/", id)
	returnedData, _, err := r.client.CreateUpdateAPIRequest(ctx, http.MethodPatch, url, bodyData, []int{200}, "eda")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API update request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}

	data.EventStreamType = types.StringValue(fmt.Sprintf("%v", returnedData["event_stream_type"]))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EdaEventStreamResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data EdaEventStreamModel

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

	url := fmt.Sprintf("event-streams/%d/", id)
	_, _, err = r.client.GenericAPIRequest(ctx, http.MethodDelete, url, nil, []int{202, 204}, "eda")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API delete request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}
}

func (r *EdaEventStreamResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
