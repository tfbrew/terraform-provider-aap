package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &EdaProjectResource{}
var _ resource.ResourceWithImportState = &EdaProjectResource{}

func NewEdaProjectResource() resource.Resource {
	return &EdaProjectResource{}
}

type EdaProjectResource struct {
	client *providerClient
}

func (r *EdaProjectResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_eda_project"
}

func (r *EdaProjectResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an EDA Project resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the EDA project.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the EDA project.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "The description of the EDA project.",
			},
			"eda_credential_id": schema.Int32Attribute{
				Optional:    true,
				Description: "The EDA credential ID associated with the EDA project.",
			},
			"organization_id": schema.Int32Attribute{
				Required:    true,
				Description: "The organization ID for the EDA project.",
			},
			"proxy": schema.StringAttribute{
				Optional:    true,
				Description: "The proxy server for the EDA project.",
			},
			"scm_branch": schema.StringAttribute{
				Optional:    true,
				Description: "The SCM branch for the EDA project.",
			},
			"scm_refspec": schema.StringAttribute{
				Optional:    true,
				Description: "The SCM refspec for the EDA project.",
			},
			"scm_type": schema.StringAttribute{
				Optional:    true,
				Description: "The SCM type for the EDA project. Currently only `git` is an option.",
				Default:     stringdefault.StaticString("git"),
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"git"}...),
				},
			},
			"signature_validation_credential_id": schema.Int32Attribute{
				Optional:    true,
				Description: "The content signature validation credential ID for the EDA project.",
			},
			"url": schema.StringAttribute{
				Description: "The SCM URL for the EDA project.",
				Required:    true,
			},
			"verify_ssl": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether to verify SSL for the SCM URL.",
				Default:     booldefault.StaticBool(true),
				Computed:    true,
			},
		},
	}
}

func (r *EdaProjectResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *EdaProjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data EdaProjectModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var bodyData EdaProjectAPIModel

	bodyData.Name = data.Name.ValueString()
	bodyData.OrganizationId = int(data.OrganizationId.ValueInt32())
	bodyData.ScmType = data.ScmType.ValueString()
	bodyData.VerifySsl = data.VerifySsl.ValueBool()
	bodyData.Url = data.Url.ValueString()

	if !(data.Description.IsNull()) {
		bodyData.Description = data.Description.ValueString()
	}

	if !(data.EdaCredentialId.IsNull()) {
		bodyData.EdaCredentialId = int(data.EdaCredentialId.ValueInt32())
	}

	if !(data.Proxy.IsNull()) {
		bodyData.Proxy = data.Proxy.ValueString()
	}

	if !(data.ScmBranch.IsNull()) {
		bodyData.ScmBranch = data.ScmBranch.ValueString()
	}

	if !(data.ScmRefSpec.IsNull()) {
		bodyData.ScmRefSpec = data.ScmRefSpec.ValueString()
	}
	if !(data.SignatureValidationCredentialId.IsNull()) {
		bodyData.SignatureValidationCredentialId = int(data.SignatureValidationCredentialId.ValueInt32())
	}

	url := "projects/"
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

func (r *EdaProjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data EdaProjectModel

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

	url := fmt.Sprintf("projects/%d/", id)
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

	var responseData EdaProjectAPIModel

	err = json.Unmarshal(body, &responseData)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to unmarshal json",
			fmt.Sprintf("bodyData: %+v.", body))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), responseData.Name)...)

	// prefer nested organization.id if present
	orgID := responseData.OrganizationId
	if responseData.Organization.Id != 0 {
		orgID = responseData.Organization.Id
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), orgID)...)

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("scm_type"), responseData.ScmType)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("verify_ssl"), responseData.VerifySsl)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("url"), responseData.Url)...)

	if !data.Description.IsNull() || responseData.Description != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("description"), responseData.Description)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !data.EdaCredentialId.IsNull() || responseData.EdaCredentialId != 0 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("eda_credential_id"), responseData.EdaCredentialId)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !data.Proxy.IsNull() || responseData.Proxy != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("proxy"), responseData.Proxy)...)
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

	if !data.ScmRefSpec.IsNull() || responseData.ScmRefSpec != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("scm_refspec"), responseData.ScmRefSpec)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !data.SignatureValidationCredentialId.IsNull() || responseData.SignatureValidationCredentialId != 0 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("signature_validation_credential_id"), responseData.SignatureValidationCredentialId)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
}

func (r *EdaProjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data EdaProjectModel

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

	var bodyData EdaProjectAPIModel

	bodyData.Name = data.Name.ValueString()
	bodyData.OrganizationId = int(data.OrganizationId.ValueInt32())
	bodyData.ScmType = data.ScmType.ValueString()
	bodyData.VerifySsl = data.VerifySsl.ValueBool()
	bodyData.Url = data.Url.ValueString()

	if !(data.Description.IsNull()) {
		bodyData.Description = data.Description.ValueString()
	}

	if !(data.EdaCredentialId.IsNull()) {
		bodyData.EdaCredentialId = int(data.EdaCredentialId.ValueInt32())
	}

	if !(data.Proxy.IsNull()) {
		bodyData.Proxy = data.Proxy.ValueString()
	}

	if !(data.ScmBranch.IsNull()) {
		bodyData.ScmBranch = data.ScmBranch.ValueString()
	}

	if !(data.ScmRefSpec.IsNull()) {
		bodyData.ScmRefSpec = data.ScmRefSpec.ValueString()
	}
	if !(data.SignatureValidationCredentialId.IsNull()) {
		bodyData.SignatureValidationCredentialId = int(data.SignatureValidationCredentialId.ValueInt32())
	}

	url := fmt.Sprintf("projects/%d/", id)
	_, _, err = r.client.CreateUpdateAPIRequest(ctx, http.MethodPatch, url, bodyData, []int{200}, "eda")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API update request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EdaProjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data EdaProjectModel

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

	url := fmt.Sprintf("projects/%d/", id)
	// 403 & 409 return code indicates project is being used by a job or had sync failures.
	// Newly create projects will be syncing and not able to be deleted immediately.
	attempts := 30
	var del_err error
	var statusCode int
	for attempt := 0; attempt < attempts; attempt++ {
		_, statusCode, del_err = r.client.GenericAPIRequest(ctx, http.MethodDelete, url, nil, []int{202, 204}, "eda")
		if del_err == nil {
			if attempt > 0 {
				resp.Diagnostics.AddWarning(
					"Retry required to delete project",
					fmt.Sprintf("Project was successfully delete after %d attempt(s).", attempt))
			}
			return
		}
		if statusCode != 403 && statusCode != 409 {
			resp.Diagnostics.AddError(
				"Error making API delete request",
				fmt.Sprintf("Error was: %s.", del_err.Error()))
			return
		}
		time.Sleep(4 * time.Second)
	}

	if del_err != nil {
		resp.Diagnostics.AddError(
			"Error making API delete request",
			fmt.Sprintf("Error after %v attempt(s) was: %s.", attempts, del_err.Error()))
		return
	}
}

func (r *EdaProjectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
