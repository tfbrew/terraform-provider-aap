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

var _ datasource.DataSource = &EdaDecisionEnvironmentDataSource{}

func NewEdaDecisionEnvironmentDataSource() datasource.DataSource {
	return &EdaDecisionEnvironmentDataSource{}
}

type EdaDecisionEnvironmentDataSource struct {
	client *providerClient
}

func (d *EdaDecisionEnvironmentDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_eda_decision_environment"
}

func (d *EdaDecisionEnvironmentDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Get decision environment datasource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:    true,
				Description: "Decision Environment ID.",
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "The name of the EDA Decision Environment.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "The description of the EDA Decision Environment.",
			},
			"eda_credential_id": schema.Int32Attribute{
				Computed:    true,
				Description: "The EDA credential ID associated with the EDA Decision Environment.",
			},
			"organization_id": schema.Int32Attribute{
				Computed:    true,
				Description: "The organization ID for the EDA Decision Environment.",
			},
			"image_url": schema.StringAttribute{
				Computed:    true,
				Description: "The image URL for the EDA Decision Environment.",
			},
			"pull_policy": schema.StringAttribute{
				Computed:    true,
				Description: "The pull policy for the EDA Decision Environment. Either `always` or `never`.",
			},
		},
	}
}

func (d *EdaDecisionEnvironmentDataSource) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.Conflicting(
			path.MatchRoot("id"),
			path.MatchRoot("name"),
		),
	}
}

func (d *EdaDecisionEnvironmentDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *EdaDecisionEnvironmentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data EdaDecisionEnvironmentModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var url string

	if !data.Id.IsNull() {
		// set url for read by id HTTP request
		id, err := strconv.Atoi(data.Id.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Can't generate read() url with Id.",
				fmt.Sprintf("Unable to convert id: %v. ", data.Id.ValueString()))
			return
		}
		url = fmt.Sprintf("decision-environments/%d/", id)
	}
	if !data.Name.IsNull() {
		// set url for read by name HTTP request
		name := urlParser.QueryEscape(data.Name.ValueString())
		url = fmt.Sprintf("decision-environments/?name=%s", name)
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

	var responseData EdaDecisionEnvironmentAPIModel

	if !data.Id.IsNull() && data.Name.IsNull() {
		err = json.Unmarshal(body, &responseData)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to unmarshal response body into object",
				fmt.Sprintf("Error =  %v.", err.Error()))
			return
		}
	}
	// If looking up by name, check that there is only one response and extract it.
	if data.Id.IsNull() && !data.Name.IsNull() {
		nameResult := struct {
			Count   int                              `json:"count"`
			Results []EdaDecisionEnvironmentAPIModel `json:"results"`
		}{}
		err = json.Unmarshal(body, &nameResult)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to unmarshal response body into result object",
				fmt.Sprintf("Error:  %v.", err.Error()))
			return
		}
		if nameResult.Count == 1 {
			responseData = nameResult.Results[0]
		} else {
			resp.Diagnostics.AddError(
				"Incorrect number of credential_types returned by name",
				fmt.Sprintf("Unable to read credential_type as API returned %v credential_types.", nameResult.Count))
			return
		}
	}

	idAsString := strconv.Itoa(responseData.Id)
	data.Id = types.StringValue(idAsString)

	if responseData.OrganizationId != 0 {
		data.OrganizationId = types.Int32Value(int32(responseData.OrganizationId))
	} else if responseData.Organization.Id != 0 {
		data.OrganizationId = types.Int32Value(int32(responseData.Organization.Id))
	}

	data.ImageUrl = types.StringValue(responseData.ImageUrl)
	data.PullPolicy = types.StringValue(responseData.PullPolicy)

	data.Name = types.StringValue(responseData.Name)

	if responseData.Description != "" {
		data.Description = types.StringValue(responseData.Description)
	}

	if responseData.EdaCredentialId != 0 {
		data.EdaCredentialId = types.Int32Value(int32(responseData.EdaCredentialId))
	} else if responseData.EdaCredential.Id != 0 {
		data.EdaCredentialId = types.Int32Value(int32(responseData.EdaCredential.Id))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}
