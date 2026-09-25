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
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var _ datasource.DataSource = &CredentialDataSource{}

func NewCredentialDataSource() datasource.DataSource {
	return &CredentialDataSource{}
}

type CredentialDataSource struct {
	client *providerClient
}

func (d *CredentialDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_credential"
}

func (d *CredentialDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Get credential datasource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Credential ID. You must specify either the `id` or `name` field, but not both.",
				Optional:    true,
			},
			"name": schema.StringAttribute{
				Description: "Credential name. You must specify either the `id` or `name` field, but not both. To lookup by name you must also include `credential_type` and `organization`. Credentials without an organization, such as personal credentials owned by a user, cannot be looked up by name; use `id` instead.",
				Optional:    true,
			},
			"description": schema.StringAttribute{
				Description: "Credential description.",
				Computed:    true,
			},
			"kind": schema.StringAttribute{
				Description: "Credential kind.",
				Computed:    true,
			},
			"organization": schema.Int32Attribute{
				Description: "ID of organization which owns this credential. Required when looking up by `name`.",
				Optional:    true,
				Computed:    true,
			},
			"team": schema.Int32Attribute{
				Description: "ID of team which owns this credential. One and only one of `organization`, `team`, or `user` must be set.",
				Computed:    true,
			},
			"user": schema.Int32Attribute{
				Description: "ID of user which owns this credential. One and only one of `organization`, `team`, or `user` must be set.",
				Computed:    true,
			},
			"credential_type": schema.Int32Attribute{
				Description: "ID of the credential type. Required when looking up by `name`.",
				Optional:    true,
				Computed:    true,
			},
			"inputs": schema.StringAttribute{
				Description: "Credential inputs. This is a JSON string representing a dictionary of inputs.",
				Computed:    true,
			},
			"inputs_as_object": schema.DynamicAttribute{
				Description: "Credential inputs as object. This is the same data as `inputs` but in object format.",
				Computed:    true,
			},
		},
	}
}

func (d *CredentialDataSource) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.ExactlyOneOf(
			path.MatchRoot("id"),
			path.MatchRoot("name"),
		),
		datasourcevalidator.RequiredTogether(
			path.MatchRoot("name"),
			path.MatchRoot("credential_type"),
			path.MatchRoot("organization"),
		),
	}
}

func (d *CredentialDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CredentialDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CredentialDataModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	query := urlParser.Values{}

	if !data.Id.IsNull() {
		id, err := strconv.Atoi(data.Id.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable convert id from string to int.",
				fmt.Sprintf("Unable to convert id: %v. ", data.Id.ValueString()))
			return
		}
		query.Set("id", strconv.Itoa(id))
	}
	if !data.Name.IsNull() {
		query.Set("name", data.Name.ValueString())
	}
	if !data.CredentialType.IsNull() {
		query.Set("credential_type", strconv.Itoa(int(data.CredentialType.ValueInt32())))
	}
	if !data.Organization.IsNull() {
		query.Set("organization", strconv.Itoa(int(data.Organization.ValueInt32())))
	}

	url := "credentials/?" + query.Encode()
	body, _, err := d.client.GenericAPIRequest(ctx, http.MethodGet, url, nil, []int{200}, "")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving datasource. The resource may not exist.",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}

	var responseData CredentialAPIModel

	countResult := struct {
		Count   int                  `json:"count"`
		Results []CredentialAPIModel `json:"results"`
	}{}

	err = json.Unmarshal(body, &countResult)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to unmarshal response body into object",
			fmt.Sprintf("Error =  %v.", err.Error()))
		return
	}
	if countResult.Count != 1 {
		resp.Diagnostics.AddError(
			"Incorrect number of credentials returned",
			fmt.Sprintf("Unable to read credential as API returned %v credentials for query %q.", countResult.Count, query.Encode()))
		return
	}
	responseData = countResult.Results[0]

	idAsString := strconv.Itoa(responseData.Id)
	data.Id = types.StringValue(idAsString)

	data.Name = types.StringValue(responseData.Name)

	if responseData.Description != "" {
		data.Description = types.StringValue(responseData.Description)
	}

	if responseData.Organization != 0 {
		data.Organization = types.Int32Value(int32(responseData.Organization))
	}

	data.Kind = types.StringValue(responseData.Kind)
	data.CredentialType = types.Int32Value(int32(responseData.CredentialType))

	jsonInputs, err := json.Marshal(responseData.Inputs)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Convert to string and print
	jsonString := string(jsonInputs)

	data.Inputs = types.StringValue(jsonString)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	var dynValue basetypes.DynamicValue
	resp.Diagnostics.Append(credentialInputApiToDynamicObject(&responseData.Inputs, &dynValue)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("inputs_as_object"), &dynValue)...)
}
