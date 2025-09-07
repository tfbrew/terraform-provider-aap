package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &RoleDefinitionDataSource{}

func NewRoleDefinitionDataSource() datasource.DataSource {
	return &RoleDefinitionDataSource{}
}

type RoleDefinitionDataSource struct {
	client *providerClient
}

func (d *RoleDefinitionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role_definition"
}

func (d *RoleDefinitionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Get role definition datasource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "A valid role definition ID. You must specify either the `id` or `name` field, but not both.",
				Optional:    true,
			},
			"name": schema.StringAttribute{
				Description: "Role definition name. You must specify either the `id` or `name` field, but not both.",
				Optional:    true,
			},
			"description": schema.StringAttribute{
				Description: "Role definition description.",
				Computed:    true,
			},
			"content_type": schema.StringAttribute{
				Description: "The type of resource this applies to.",
				Computed:    true,
			},
			"permissions": schema.SetAttribute{
				Description: "List of permissions to include in the role definition.",
				ElementType: types.StringType,
				Computed:    true,
			},
		},
	}
}

func (d *RoleDefinitionDataSource) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
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

func (d *RoleDefinitionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	configureData, ok := req.ProviderData.(*providerClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *providerClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = configureData
}

func (d *RoleDefinitionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RoleDefinitionModel

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
		url = fmt.Sprintf("role_definitions/?id=%d", id)
	}
	if !data.Name.IsNull() {
		url = fmt.Sprintf("role_definitions/?name=%s", data.Name.ValueString())
	}

	body, statusCode, err := d.client.GenericAPIRequest(ctx, http.MethodGet, url, nil, []int{200, 404}, "")
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

	countResult := struct {
		Count   int                      `json:"count"`
		Results []RoleDefinitionAPIModel `json:"results"`
	}{}

	err = json.Unmarshal(body, &countResult)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to unmarshal response body into object",
			fmt.Sprintf("Error:  %v.", err.Error()))
		return
	}
	if countResult.Count == 1 {
		responseData := countResult.Results[0]
		data.Id = types.StringValue(fmt.Sprintf("%v", responseData.Id))
		data.Name = types.StringValue(responseData.Name)
		data.ContentType = types.StringValue(responseData.ContentType)
		data.Permissions = make([]types.String, len(responseData.Permissions))
		for i, perm := range responseData.Permissions {
			data.Permissions[i] = types.StringValue(perm)
		}
		if responseData.Description != "" {
			data.Description = types.StringValue(responseData.Description)
		}
	} else {
		resp.Diagnostics.AddError(
			"Incorrect number of role definitions returned",
			fmt.Sprintf("Unable to read role definition as API returned %v results.", countResult.Count))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
