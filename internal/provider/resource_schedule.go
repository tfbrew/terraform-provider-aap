package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var _ resource.Resource = &ScheduleResource{}
var _ resource.ResourceWithImportState = &ScheduleResource{}

func NewScheduleResource() resource.Resource {
	return &ScheduleResource{}
}

type ScheduleResource struct {
	client *providerClient
}

func (r *ScheduleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_schedule"
}

func (r *ScheduleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `Manage an Automation Controller schedule.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Schedule ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Schedule name.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Schedule description.",
				Optional:    true,
			},
			"unified_job_template": schema.Int32Attribute{
				Description: "Job template id for schedule.",
				Required:    true,
			},
			"rrule": schema.StringAttribute{
				Description: "Schedule rrule (i.e. `DTSTART;TZID=America/Chicago:20250124T090000 RRULE:INTERVAL=1;FREQ=WEEKLY;BYDAY=TU`.",
				Required:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Schedule enabled (defaults true).",
				Optional:    true,
				Default:     booldefault.StaticBool(true),
				Computed:    true,
			},
			"inventory": schema.Int32Attribute{
				Description: "Inventory id for schedule - for providing prompt values",
				Optional:    true,
			},
			"limit": schema.StringAttribute{
				Description: "Limit - for providing prompt values.",
				Optional:    true,
			},
			"forks": schema.Int32Attribute{
				Description: "Forks for schedule - for providing prompt values",
				Optional:    true,
			},
			"job_slice_count": schema.Int32Attribute{
				Description: "Job Slice Count for schedule - for providing prompt values",
				Optional:    true,
			},
			"scm_branch": schema.StringAttribute{
				Description: "SCM Branch - for providing prompt values.",
				Optional:    true,
			},
			"job_tags": schema.StringAttribute{
				Description: "Job Tags - for providing prompt values.",
				Optional:    true,
			},
			"skip_tags": schema.StringAttribute{
				Description: "Skip Tags - for providing prompt values.",
				Optional:    true,
			},
			"diff_mode": schema.BoolAttribute{
				Description: "Diff Mode - for providing prompt values.",
				Optional:    true,
			},
			"verbosity": schema.Int32Attribute{
				Description: "Verbosity - for providing prompt values",
				Optional:    true,
			},
			"execution_environment": schema.Int32Attribute{
				Description: "Execution Environment id - for providing prompt values",
				Optional:    true,
			},
			"timeout": schema.Int32Attribute{
				Description: "Timeout - for providing prompt values",
				Optional:    true,
			},
			"extra_data": schema.DynamicAttribute{
				Description: "This field expects an object of key value pairs. If the value for a particular key is a list of strings, wrap the value in `tolist([...])` in your terraform configuration code. See the example schedule for more details.",
				Optional:    true,
			},
		},
	}
}

func (r *ScheduleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ScheduleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ScheduleModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var bodyData ScheduleAPIModel

	bodyData.Name = data.Name.ValueString()
	bodyData.UnifiedJobTemplate = int(data.UnifiedJobTemplate.ValueInt32())
	bodyData.Rrule = data.Rrule.ValueString()
	bodyData.Enabled = data.Enabled.ValueBool()
	if !(data.Description.IsNull()) {
		bodyData.Description = data.Description.ValueString()
	}

	if !(data.Inventory.IsNull()) {
		bodyData.Inventory = int(data.Inventory.ValueInt32())
	}
	if !(data.Limit.IsNull()) {
		bodyData.Limit = data.Limit.ValueString()
	}
	if !(data.Forks.IsNull()) {
		bodyData.Forks = int(data.Forks.ValueInt32())
	}
	if !(data.JobSliceCount.IsNull()) {
		bodyData.JobSliceCount = int(data.JobSliceCount.ValueInt32())
	}
	if !(data.Verbosity.IsNull()) {
		bodyData.Verbosity = int(data.Verbosity.ValueInt32())
	}
	if !(data.ExecutionEnvironment.IsNull()) {
		bodyData.ExecutionEnvironment = int(data.ExecutionEnvironment.ValueInt32())
	}
	if !(data.Timeout.IsNull()) {
		bodyData.Timeout = int(data.Timeout.ValueInt32())
	}
	if !(data.ScmBranch.IsNull()) {
		bodyData.ScmBranch = data.ScmBranch.ValueString()
	}
	if !(data.JobTags.IsNull()) {
		bodyData.JobTags = data.JobTags.ValueString()
	}
	if !(data.SkipTags.IsNull()) {
		bodyData.SkipTags = data.SkipTags.ValueString()
	}
	if !(data.DiffMode.IsNull()) {
		bodyData.DiffMode = data.DiffMode.ValueBool()
	}

	if !data.ExtraData.IsUnderlyingValueNull() && !data.ExtraData.IsNull() {
		extraDataMap := make(map[string]any)

		switch val := data.ExtraData.UnderlyingValue().(type) {
		case types.Object:

			for key, v := range val.Attributes() {
				switch v := v.(type) {
				case basetypes.StringValue:
					extraDataMap[key] = v.ValueString()

				case basetypes.NumberValue:
					f64, _ := v.ValueBigFloat().Float64()
					extraDataMap[key] = f64

				case basetypes.ListValue:

					beforeStringConversion := v.Elements()
					stringSlice := make([]string, len(beforeStringConversion))
					for i, elem := range beforeStringConversion {
						strElem, ok := elem.(types.String)
						if !ok {
							resp.Diagnostics.AddError(
								"extra data list contains non-string element",
								fmt.Sprintf("extra data key '%s' has a non-string element of type: %T", key, elem),
							)
							return
						}
						stringSlice[i] = strElem.ValueString()
					}
					extraDataMap[key] = stringSlice

				default:
					resp.Diagnostics.AddError(
						"extra data value specified is invalid type create",
						fmt.Sprintf("extra data key '%s' has an unexpected type: %T. Wrap lists in tolist([...]) for the list element within extra data.", key, v),
					)
					return
				}
			}
		default:
			resp.Diagnostics.AddError("Extra data type invalid (create)", "The extra_data should be a types.Object.")
			return
		}

		bodyData.ExtraData = extraDataMap
	}

	url := "schedules/"
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

func (r *ScheduleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ScheduleModel

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

	url := fmt.Sprintf("schedules/%d/", id)
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

	var responseData ScheduleAPIModel

	err = json.Unmarshal(body, &responseData)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to unmarshal json",
			fmt.Sprintf("bodyData: %+v.", body))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), responseData.Name)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("unified_job_template"), responseData.UnifiedJobTemplate)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("rrule"), responseData.Rrule)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("enabled"), responseData.Enabled)...)

	if !data.Description.IsNull() || responseData.Description != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("description"), responseData.Description)...)
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
	if !data.ScmBranch.IsNull() || responseData.ScmBranch != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("scm_branch"), responseData.ScmBranch)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	if !data.JobTags.IsNull() || responseData.JobTags != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("job_tags"), responseData.JobTags)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	if !data.DiffMode.IsNull() || responseData.DiffMode {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("diff_mode"), responseData.DiffMode)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	if !data.DiffMode.IsNull() || responseData.DiffMode {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("diff_mode"), responseData.DiffMode)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	if !data.Inventory.IsNull() || responseData.Inventory != 0 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("inventory"), responseData.Inventory)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	if !data.Timeout.IsNull() || responseData.Timeout != 0 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("timeout"), responseData.Timeout)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	if !data.Forks.IsNull() || responseData.Forks != 0 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("forks"), responseData.Forks)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	if !data.JobSliceCount.IsNull() || responseData.JobSliceCount != 0 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("job_slice_count"), responseData.JobSliceCount)...)
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
	if !data.ExecutionEnvironment.IsNull() || responseData.ExecutionEnvironment != 0 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("execution_environment"), responseData.ExecutionEnvironment)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !data.ExtraData.IsUnderlyingValueNull() && !data.ExtraData.IsNull() {

		extraDataValue := data.ExtraData.UnderlyingValue()

		// convert state to map[string]any
		currExtraDataState := make(map[string]any)

		switch val := extraDataValue.(type) {
		case types.Object:
			for key, v := range val.Attributes() {
				switch v := v.(type) {
				case basetypes.StringValue:
					currExtraDataState[key] = v.ValueString()

				case basetypes.NumberValue:
					f64, _ := v.ValueBigFloat().Float64()
					currExtraDataState[key] = f64

				case basetypes.ListValue:

					beforeStringConversion := v.Elements()
					stringSlice := make([]string, len(beforeStringConversion))
					for i, elem := range beforeStringConversion {
						strElem, ok := elem.(types.String)
						if !ok {
							resp.Diagnostics.AddError(
								"extra data tuple contains non-string element",
								fmt.Sprintf("extra data key '%s' has a non-string element of type: %T", key, elem),
							)
							return
						}
						stringSlice[i] = strElem.ValueString()
					}
					currExtraDataState[key] = stringSlice

				default:
					resp.Diagnostics.AddError(
						"extra data value specified is invalid type",
						fmt.Sprintf("extra data key '%s' has an unexpected type: %T", key, v),
					)
					return
				}
			}

			var dynValue basetypes.DynamicValue
			resp.Diagnostics.Append(extraDataApiToDynamicObject(ctx, responseData.ExtraData, &dynValue)...)
			if resp.Diagnostics.HasError() {
				return
			}

			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("extra_data"), &dynValue)...)
			if resp.Diagnostics.HasError() {
				return
			}

		default:
			resp.Diagnostics.AddError("extra data value specified is invalid type", "extra data must be an object.")
			return
		}
	}

}

func (r *ScheduleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ScheduleModel

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

	var bodyData ScheduleAPIModel

	bodyData.Name = data.Name.ValueString()
	bodyData.UnifiedJobTemplate = int(data.UnifiedJobTemplate.ValueInt32())
	bodyData.Rrule = data.Rrule.ValueString()
	bodyData.Enabled = data.Enabled.ValueBool()

	if !(data.Description.IsNull()) {
		bodyData.Description = data.Description.ValueString()
	}
	if !(data.Inventory.IsNull()) {
		bodyData.Inventory = int(data.Inventory.ValueInt32())
	}
	if !(data.Limit.IsNull()) {
		bodyData.Limit = data.Limit.ValueString()
	}
	if !(data.Forks.IsNull()) {
		bodyData.Forks = int(data.Forks.ValueInt32())
	}
	if !(data.JobSliceCount.IsNull()) {
		bodyData.JobSliceCount = int(data.JobSliceCount.ValueInt32())
	}
	if !(data.Verbosity.IsNull()) {
		bodyData.Verbosity = int(data.Verbosity.ValueInt32())
	}
	if !(data.ExecutionEnvironment.IsNull()) {
		bodyData.ExecutionEnvironment = int(data.ExecutionEnvironment.ValueInt32())
	}
	if !(data.Timeout.IsNull()) {
		bodyData.Timeout = int(data.Timeout.ValueInt32())
	}
	if !(data.ScmBranch.IsNull()) {
		bodyData.ScmBranch = data.ScmBranch.ValueString()
	}
	if !(data.JobTags.IsNull()) {
		bodyData.JobTags = data.JobTags.ValueString()
	}
	if !(data.SkipTags.IsNull()) {
		bodyData.SkipTags = data.SkipTags.ValueString()
	}
	if !(data.DiffMode.IsNull()) {
		bodyData.DiffMode = data.DiffMode.ValueBool()
	}

	if !data.ExtraData.IsUnderlyingValueNull() && !data.ExtraData.IsNull() {
		extraDataMap := make(map[string]any)

		switch val := data.ExtraData.UnderlyingValue().(type) {

		case types.Object:
			for key, v := range val.Attributes() {

				switch v := v.(type) {
				case basetypes.StringValue:
					extraDataMap[key] = v.ValueString()

				case basetypes.NumberValue:
					f64, _ := v.ValueBigFloat().Float64()
					extraDataMap[key] = f64

				case basetypes.ListValue:

					beforeStringConversion := v.Elements()
					stringSlice := make([]string, len(beforeStringConversion))
					for i, elem := range beforeStringConversion {
						strElem, ok := elem.(types.String)
						if !ok {
							resp.Diagnostics.AddError(
								"extra data tuple contains non-string element",
								fmt.Sprintf("extra data key '%s' has a non-string element of type: %T", key, elem),
							)
							return
						}
						stringSlice[i] = strElem.ValueString()
					}
					extraDataMap[key] = stringSlice
				default:
					resp.Diagnostics.AddError(
						"extra data value specified is invalid type",
						fmt.Sprintf("extra data key '%s' has an unexpected type: %T", key, v),
					)
					return
				}
			}
		default:
			resp.Diagnostics.AddError("Extra Data type invalid", "The extra data should be a types.Object.")
			return
		}

		bodyData.ExtraData = extraDataMap
	}

	url := fmt.Sprintf("schedules/%d/", id)
	_, _, err = r.client.CreateUpdateAPIRequest(ctx, http.MethodPut, url, bodyData, []int{200}, "")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API update request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ScheduleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ScheduleModel

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

	url := fmt.Sprintf("schedules/%d/", id)
	_, _, err = r.client.GenericAPIRequest(ctx, http.MethodDelete, url, nil, []int{202, 204}, "")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API delete request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}
}

func (r *ScheduleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
