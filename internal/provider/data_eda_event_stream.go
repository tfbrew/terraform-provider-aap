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

func NewEdaEventStreamDataSource() datasource.DataSource {
	return &EdaEventStreamDataSource{}
}

type EdaEventStreamDataSource struct {
	client *providerClient
}

func (d *EdaEventStreamDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_eda_event_stream"
}

func (d *EdaEventStreamDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Get EDA event stream datasource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:    true,
				Description: "The ID of the EDA event stream. Either specify `id` or `name` but not both.",
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "The name of the EDA event stream. Either specify `id` or `name` but not both.",
			},
			"additional_data_headers": schema.StringAttribute{
				Computed:    true,
				Description: "The additional http headers which will be added to the event data. The headers are comma delimited.",
			},
			"eda_credential_id": schema.Int32Attribute{
				Computed:    true,
				Description: "The EDA credential ID associated with the EDA Event Stream.",
			},
			"organization_id": schema.Int32Attribute{
				Computed:    true,
				Description: "The organization ID for the EDA Event Stream.",
			},
			"test_mode": schema.BoolAttribute{
				Description: "Enable test mode.",
				Computed:    true,
			},
			"uuid": schema.StringAttribute{
				Computed:    true,
				Description: "The uuid for the EDA Event Stream.",
			},
			"url": schema.StringAttribute{
				Computed:    true,
				Description: "The url for the EDA Event Stream.",
			},
			"event_stream_type": schema.StringAttribute{
				Computed:    true,
				Description: "The event stream type for the EDA Event Stream based on credential type.",
			},
		},
	}
}

func (d *EdaEventStreamDataSource) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
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

func (d *EdaEventStreamDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *EdaEventStreamDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data EdaEventStreamModel

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
		url = fmt.Sprintf("event-streams/%d/", id)
	}
	if !data.Name.IsNull() {
		url = fmt.Sprintf("event-streams/?name=%s", urlParser.QueryEscape(data.Name.ValueString()))
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

	var responseData EdaEventStreamAPIModel

	if data.Id.IsNull() {
		countResult := struct {
			Count   int                      `json:"count"`
			Results []EdaEventStreamAPIModel `json:"results"`
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
				"Incorrect number of event streams returned",
				fmt.Sprintf("Unable to read event stream as API returned %v event streams.", countResult.Count))
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

	// prefer nested organization.id if present
	edaCredID := responseData.EdaCredentialId
	if responseData.EdaCredential.Id != 0 {
		edaCredID = responseData.EdaCredential.Id
	}

	data.Name = types.StringValue(responseData.Name)
	data.OrganizationId = types.Int32Value(int32(orgID))
	data.EdaCredentialId = types.Int32Value(int32(edaCredID))
	data.TestMode = types.BoolValue(responseData.TestMode)
	data.Uuid = types.StringValue(responseData.Uuid)
	data.Url = types.StringValue(responseData.Url)
	data.EventStreamType = types.StringValue(responseData.EventStreamType)
	data.AdditionalDataHeaders = types.StringValue(responseData.AdditionalDataHeaders)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
