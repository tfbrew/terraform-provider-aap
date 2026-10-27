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
	"github.com/tfbrew/terraform-provider-aap/internal/configprefix"
)

var _ datasource.DataSource = &OrganizationDataSource{}

func NewOrganizationDataSource() datasource.DataSource {
	return &OrganizationDataSource{}
}

type OrganizationDataSource struct {
	client *providerClient
}

func (d *OrganizationDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization"
}

func (d *OrganizationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Get organization datasource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: configprefix.OrgDataSourceIdDescription,
				Optional:    true,
			},
			// AAP-AWX-DIFF-BEGIN: aap also exposes the gateway and EDA organization IDs.
			"gateway_id": schema.Int32Attribute{
				Description: "Organization ID in the gateway API",
				Computed:    true,
			},
			"eda_id": schema.Int32Attribute{
				Description: "Organization ID in the EDA API",
				Computed:    true,
			},
			// AAP-AWX-DIFF-END
			"name": schema.StringAttribute{
				Description: "Organization name",
				Optional:    true,
			},
			"description": schema.StringAttribute{
				Description: "Organization description.",
				Computed:    true,
			},
			"max_hosts": schema.Int32Attribute{
				Description: "Maximum number of hosts allowed to be managed by this organization.",
				Computed:    true,
			},
			// AAP-AWX-DIFF-BEGIN: aap also exposes opa_query_path; awx exposes default_environment.
			"opa_query_path": schema.StringAttribute{
				Description: "The query path of the OPA policy evaluated before jobs in this organization run.",
				Computed:    true,
			},
			// AAP-AWX-DIFF-END
		},
	}
}

func (d OrganizationDataSource) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.ExactlyOneOf(
			path.MatchRoot("id"),
			path.MatchRoot("name"),
		),
	}
}

func (d *OrganizationDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *OrganizationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OrganizationModel

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
				"Unable convert id from string to int.",
				fmt.Sprintf("Unable to convert id: %v. ", data.Id.ValueString()))
			return
		}
		url = fmt.Sprintf("organizations/%d/", id)
	}
	if !data.Name.IsNull() {
		// set url for read by name HTTP request
		name := urlParser.QueryEscape(data.Name.ValueString())
		url = fmt.Sprintf("organizations/?name=%s", name)
	}

	// let function know we're a data source instead of a resource
	var contextKey contextKey = "dataSource"

	ctx = context.WithValue(ctx, contextKey, true)
	body, _, err := d.client.GenericAPIRequest(ctx, http.MethodGet, url, nil, []int{200}, "")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving datasource. The resource may not exist.",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}

	var responseData OrganizationAPIModel

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
			Count   int                    `json:"count"`
			Results []OrganizationAPIModel `json:"results"`
		}{}
		err = json.Unmarshal(body, &nameResult)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to unmarshal response body into object",
				fmt.Sprintf("Error:  %v.", err.Error()))
			return
		}
		if nameResult.Count == 1 {
			responseData = nameResult.Results[0]
		} else {
			resp.Diagnostics.AddError(
				"Incorrect number of organizations returned by name",
				fmt.Sprintf("Unable to read organization as API returned %v organizations.", nameResult.Count))
			return
		}
	}

	idAsString := strconv.Itoa(responseData.Id)
	data.Id = types.StringValue(idAsString)

	data.Name = types.StringValue(responseData.Name)

	if responseData.Description != "" {
		data.Description = types.StringValue(responseData.Description)
	}

	data.MaxHosts = types.Int32Value(int32(responseData.MaxHosts))
	// AAP-AWX-DIFF-BEGIN: aap also reads opa_query_path and looks up the gateway and EDA organization IDs; awx reads default_environment.
	data.OpaQueryPath = types.StringValue(responseData.OpaQueryPath)

	// now get the EDA ID by querying the eda endpoint

	// the org may have been created in gateway moments ago, so wait for it to be synced to EDA.
	// If the eda endpoint doesn't exist (404), just continue without error and leave the EDA ID empty.
	edaId, found, err := lookupOrgIdByName(ctx, d.client, data.Name.ValueString(), "eda", true)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to look up organization in EDA API",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}
	if found {
		data.EdaId = types.Int32Value(int32(edaId))
	}

	// now get the Gateway ID by querying the gateway endpoint

	gatewayId, found, err := lookupOrgIdByName(ctx, d.client, data.Name.ValueString(), "gateway", false)
	if err == nil && !found {
		err = fmt.Errorf("gateway endpoint returned 404")
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to look up organization in gateway API",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}
	data.GatewayId = types.Int32Value(int32(gatewayId))
	// AAP-AWX-DIFF-END

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
