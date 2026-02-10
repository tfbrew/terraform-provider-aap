package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var _ datasource.DataSource = &EdaCredentialDataSource{}

func NewEdaCredentialDataSource() datasource.DataSource {
	return &EdaCredentialDataSource{}
}

type EdaCredentialDataSource struct {
	client *providerClient
}

func (d *EdaCredentialDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_eda_credential"
}

func (d *EdaCredentialDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Get credential datasource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Credential ID.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Credential name.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Credential description.",
				Computed:    true,
			},
			"organization_id": schema.Int32Attribute{
				Description: "ID of organization which owns this credential.",
				Computed:    true,
			},
			"credential_type_id": schema.Int32Attribute{
				Description: "ID of the credential type.",
				Computed:    true,
			},
			"inputs": schema.StringAttribute{
				Description: "Specify a string by using using `jsonencode()` to encode similar data as as string in state. Specify alphabetically when using the second method.",
				Computed:    true,
			},
			"inputs_as_object": schema.DynamicAttribute{
				Description: "Credential inputs as object. This is the same data as `inputs` but in object format.",
				Computed:    true,
			},
		},
	}
}

func (d *EdaCredentialDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *EdaCredentialDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data EdaCredentialDataModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var url string

	id, err := strconv.Atoi(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable convert id from string to int.",
			fmt.Sprintf("Unable to convert id: %v. ", data.Id.ValueString()))
		return
	}

	url = fmt.Sprintf("eda-credentials/%d/", id)
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

	var responseData EdaCredentialAPIModel

	err = json.Unmarshal(body, &responseData)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to unmarshal response body into object",
			fmt.Sprintf("Error =  %v.", err.Error()))
		return
	}

	idAsString := strconv.Itoa(responseData.Id)
	data.Id = types.StringValue(idAsString)

	data.Name = types.StringValue(responseData.Name)

	if responseData.Description != "" {
		data.Description = types.StringValue(responseData.Description)
	}

	if responseData.Organization.Id != 0 {
		data.OrganizationId = types.Int32Value(int32(responseData.Organization.Id))
	}

	data.CredentialTypeId = types.Int32Value(int32(responseData.CredentialTypeId))

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
