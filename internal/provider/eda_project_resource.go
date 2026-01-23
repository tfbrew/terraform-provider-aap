package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &EdaProjectResource{}
	_ resource.ResourceWithConfigure   = &EdaProjectResource{}
	_ resource.ResourceWithImportState = &EdaProjectResource{}
)

func NewEdaProjectResource() resource.Resource {
	return &EdaProjectResource{}
}

type EdaProjectResource struct {
	client *providerClient
}

type EdaProjectResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	URL            types.String `tfsdk:"url"`
	SCMBranch      types.String `tfsdk:"scm_branch"`
	OrganizationID types.Int64  `tfsdk:"organization_id"`
	Proxy          types.String `tfsdk:"proxy"`
}

type EdaProjectAPIModel struct {
	ID             int64  `json:"id,omitempty"`
	Name           string `json:"name"`
	Description    string `json:"description,omitempty"`
	URL            string `json:"url"`
	SCMBranch      string `json:"scm_branch,omitempty"`
	OrganizationID int64  `json:"organization_id"`
	Proxy          string `json:"proxy,omitempty"`
}

type EdaProjectListResponse struct {
	Count   int                  `json:"count"`
	Results []EdaProjectAPIModel `json:"results"`
}

func (r *EdaProjectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_eda_project"
}

func (r *EdaProjectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"url": schema.StringAttribute{
				Required:    true,
				Description: "The SCM URL for the EDA project.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"scm_branch": schema.StringAttribute{
				Optional:    true,
				Description: "The SCM branch for the EDA project.",
			},
			"organization_id": schema.Int64Attribute{
				Required:    true,
				Description: "The organization ID for the EDA project.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"proxy": schema.StringAttribute{
				Optional:    true,
				Description: "The proxy server for the EDA project.",
			},
		},
	}
}

func (r *EdaProjectResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*providerClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *providerClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *EdaProjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan EdaProjectResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	requestBody, diags := plan.generateRequestBody()
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectsURL := "api/eda/v1/projects/"
	requestData := requestBody
	createResponseBody, _, err := r.client.CreateUpdateAPIRequest(ctx, http.MethodPost, projectsURL, json.RawMessage(requestData), []int{http.StatusCreated}, "eda")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating EDA project",
			fmt.Sprintf("Could not create EDA project: %s", err.Error()),
		)
		return
	}

	responseBytes, err := json.Marshal(createResponseBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error marshaling response",
			fmt.Sprintf("Could not marshal response: %s", err.Error()),
		)
		return
	}

	resp.Diagnostics.Append(plan.parseHTTPResponse(responseBytes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *EdaProjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state EdaProjectResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectsURL := fmt.Sprintf("api/eda/v1/projects/?name=%s", state.Name.ValueString())

	readResponseBody, statusCode, err := r.client.GenericAPIRequest(ctx, http.MethodGet, projectsURL, nil, []int{http.StatusOK, http.StatusNotFound}, "eda")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading EDA project",
			fmt.Sprintf("Could not read EDA project: %s", err.Error()),
		)
		return
	}

	if statusCode == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	var listResponse EdaProjectListResponse
	err = json.Unmarshal(readResponseBody, &listResponse)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error parsing JSON response from AAP",
			fmt.Sprintf("Unable to parse EDA project list response: %s", err.Error()),
		)
		return
	}

	if listResponse.Count == 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	if listResponse.Count > 1 {
		resp.Diagnostics.AddError(
			"Multiple EDA Projects found",
			fmt.Sprintf("Expected 1 project with name %s, found %d", state.Name.ValueString(), listResponse.Count),
		)
		return
	}

	project := listResponse.Results[0]
	diags := state.parseAPIModel(&project)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *EdaProjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan EdaProjectResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	requestBody, diags := plan.generateRequestBody()
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectURL := fmt.Sprintf("api/eda/v1/projects/%s/", plan.ID.ValueString())
	requestData := requestBody
	updateResponseBody, _, err := r.client.CreateUpdateAPIRequest(ctx, http.MethodPatch, projectURL, json.RawMessage(requestData), []int{http.StatusOK}, "eda")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating EDA project",
			fmt.Sprintf("Could not update EDA project: %s", err.Error()),
		)
		return
	}

	responseBytes, err := json.Marshal(updateResponseBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error marshaling response",
			fmt.Sprintf("Could not marshal response: %s", err.Error()),
		)
		return
	}

	resp.Diagnostics.Append(plan.parseHTTPResponse(responseBytes)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *EdaProjectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id := req.ID

	var state EdaProjectResourceModel
	state.ID = types.StringValue(id)

	projectURL := fmt.Sprintf("api/eda/v1/projects/%s/", id)
	readResponseBody, _, err := r.client.GenericAPIRequest(ctx, http.MethodGet, projectURL, nil, []int{http.StatusOK}, "eda")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading EDA project",
			fmt.Sprintf("Could not read EDA project: %s", err.Error()),
		)
		return
	}

	diags := state.parseHTTPResponse(readResponseBody)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *EdaProjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state EdaProjectResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectURL := fmt.Sprintf("api/eda/v1/projects/%s/", state.ID.ValueString())
	_, statusCode, err := r.client.GenericAPIRequest(ctx, http.MethodDelete, projectURL, nil, []int{202, 204}, "eda")
	if statusCode == 404 {
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting EDA project",
			fmt.Sprintf("Could not delete EDA project: %s", err.Error()),
		)
		return
	}
}

func (r *EdaProjectResourceModel) generateRequestBody() ([]byte, diag.Diagnostics) {
	project := EdaProjectAPIModel{
		Name:           r.Name.ValueString(),
		Description:    r.Description.ValueString(),
		URL:            r.URL.ValueString(),
		SCMBranch:      r.SCMBranch.ValueString(),
		OrganizationID: r.OrganizationID.ValueInt64(),
		Proxy:          r.Proxy.ValueString(),
	}

	jsonBody, err := json.Marshal(project)
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError(
			"Error marshaling request body",
			fmt.Sprintf("Could not generate request body for EDA project resource, unexpected error: %s", err.Error()),
		)
		return nil, diags
	}

	return jsonBody, nil
}

func (r *EdaProjectResourceModel) parseHTTPResponse(body []byte) diag.Diagnostics {
	var apiProject EdaProjectAPIModel
	err := json.Unmarshal(body, &apiProject)
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError("Error parsing JSON response from AAP", err.Error())
		return diags
	}

	return r.parseAPIModel(&apiProject)
}

func (r *EdaProjectResourceModel) parseAPIModel(apiProject *EdaProjectAPIModel) diag.Diagnostics {
	r.ID = types.StringValue(fmt.Sprintf("%d", apiProject.ID))
	r.Name = types.StringValue(apiProject.Name)
	if apiProject.Description != "" {
		r.Description = types.StringValue(apiProject.Description)
	} else {
		r.Description = types.StringNull()
	}
	r.URL = types.StringValue(apiProject.URL)
	if apiProject.SCMBranch != "" {
		r.SCMBranch = types.StringValue(apiProject.SCMBranch)
	} else {
		r.SCMBranch = types.StringNull()
	}
	r.OrganizationID = types.Int64Value(apiProject.OrganizationID)
	if apiProject.Proxy != "" {
		r.Proxy = types.StringValue(apiProject.Proxy)
	} else {
		r.Proxy = types.StringNull()
	}

	return nil
}
