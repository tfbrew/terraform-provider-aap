package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	urlParser "net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ProjectDataSource{}

func NewEdaProjectDataSource() datasource.DataSource {
	return &EdaProjectDataSource{}
}

type EdaProjectDataSource struct {
	client *providerClient
}

func (d *EdaProjectDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_eda_project"
}

func (d *EdaProjectDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Get project datasource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:    true,
				Description: "The ID of the EDA project. Either specify `id` or `name` but not both.",
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "The name of the EDA project. Either specify `id` or `name` but not both.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "The description of the EDA project.",
			},
			"eda_credential_id": schema.Int32Attribute{
				Computed:    true,
				Description: "The EDA credential ID associated with the EDA project.",
			},
			"organization_id": schema.Int32Attribute{
				Computed:    true,
				Description: "The organization ID for the EDA project.",
			},
			"proxy": schema.StringAttribute{
				Computed:    true,
				Description: "The proxy server for the EDA project.",
			},
			"scm_branch": schema.StringAttribute{
				Computed:    true,
				Description: "The SCM branch for the EDA project.",
			},
			"scm_refspec": schema.StringAttribute{
				Computed:    true,
				Description: "The SCM refspec for the EDA project.",
			},
			"scm_type": schema.StringAttribute{
				Computed:    true,
				Description: "The SCM type for the EDA project. Currently only `git` is an option.",
			},
			"signature_validation_credential_id": schema.Int32Attribute{
				Computed:    true,
				Description: "The content signature validation credential ID for the EDA project.",
			},
			"url": schema.StringAttribute{
				Computed:    true,
				Description: "The SCM URL for the EDA project.",
			},
			"verify_ssl": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether to verify SSL for the SCM URL.",
			},
		},
	}
}

func (d *EdaProjectDataSource) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.Conflicting(
			path.MatchRoot("id"),
			path.MatchRoot("name"),
		),
		datasourcevalidator.AtLeastOneOf(
			path.MatchRoot("id"),
			path.MatchRoot("name"),
		),
	}
}

func (d *EdaProjectDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	configureData, ok := req.ProviderData.(*providerClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = configureData
}

func (d *EdaProjectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data EdaProjectModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var url string

	if !data.Id.IsNull() {
		id, err := strconv.Atoi(data.Id.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable convert id from string to int.",
				fmt.Sprintf("Unable to convert id: %v. ", data.Id.ValueString()))
			return
		}
		url = fmt.Sprintf("projects/%d/", id)
	}
	if !data.Name.IsNull() {
		url = fmt.Sprintf("projects/?name=%s", urlParser.QueryEscape(data.Name.ValueString()))
	}

	body, statusCode, err := d.client.GenericAPIRequest(ctx, http.MethodGet, url, nil, []int{200, 404}, "eda")
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

	if data.Id.IsNull() {
		countResult := struct {
			Count   int                  `json:"count"`
			Results []EdaProjectAPIModel `json:"results"`
		}{}

		err = json.Unmarshal(body, &countResult)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to unmarshal response body into object",
				fmt.Sprintf("Error:  %v.", err.Error()))
			return
		}
		if countResult.Count == 1 {
			responseData = countResult.Results[0]
		} else {
			resp.Diagnostics.AddError(
				"Incorrect number of projects returned",
				fmt.Sprintf("Unable to read project as API returned %v projects.", countResult.Count))
			return
		}
	} else {
		err = json.Unmarshal(body, &responseData)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to unmarshal response body into object",
				fmt.Sprintf("Error =  %v.", err.Error()))
			return
		}
	}

	idAsString := strconv.Itoa(responseData.Id)
	data.Id = types.StringValue(idAsString)

	// prefer nested organization.id if present
	orgID := responseData.OrganizationId
	if responseData.Organization.Id != 0 {
		orgID = responseData.Organization.Id
	}

	data.Name = types.StringValue(responseData.Name)
	data.OrganizationId = types.Int32Value(int32(orgID))
	data.ScmType = types.StringValue(responseData.ScmType)
	data.Url = types.StringValue(responseData.Url)
	data.VerifySsl = types.BoolValue(responseData.VerifySsl)

	if responseData.Description != "" {
		data.Description = types.StringValue(responseData.Description)
	}

	if responseData.EdaCredentialId != 0 {
		data.EdaCredentialId = types.Int32Value(int32(responseData.EdaCredentialId))
	}

	if responseData.Proxy != "" {
		data.Proxy = types.StringValue(responseData.Proxy)
	}

	if responseData.ScmBranch != "" {
		data.ScmBranch = types.StringValue(responseData.ScmBranch)
	}

	if responseData.ScmRefSpec != "" {
		data.ScmRefSpec = types.StringValue(responseData.ScmRefSpec)
	}

	if responseData.SignatureValidationCredentialId != 0 {
		data.SignatureValidationCredentialId = types.Int32Value(int32(responseData.SignatureValidationCredentialId))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
