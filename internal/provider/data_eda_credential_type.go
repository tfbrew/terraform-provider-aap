package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	urlParser "net/url"
	"reflect"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &EdaCredentialTypeDataSource{}

func NewEdaCredentialTypeDataSource() datasource.DataSource {
	return &EdaCredentialTypeDataSource{}
}

type EdaCredentialTypeDataSource struct {
	client *providerClient
}

func (d *EdaCredentialTypeDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_eda_credential_type"
}

func (d *EdaCredentialTypeDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Get eda_credential_type datasource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Crdential type ID.",
				Optional:    true,
			},
			"name": schema.StringAttribute{
				Description: "Credential type name.",
				Optional:    true,
			},
			"description": schema.StringAttribute{
				Description: "Credential type description.",
				Computed:    true,
			},
			"inputs": schema.StringAttribute{
				Computed:    true,
				Description: "Enter inputs using JSON syntax wrapped with `jsonencode()`. Must be alphabetized by key. Refer to the Ansible Controller documentation for example syntax.",
			},
			"injectors": schema.StringAttribute{
				Computed:    true,
				Description: "Enter injectors using either JSON syntax with `jsonencode()`. Must be alphabetized by key. Refer to the Ansible Controller documentation for example syntax.",
			},
			"kind": schema.StringAttribute{
				Computed: true,
			},
			"namespace": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d EdaCredentialTypeDataSource) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.Conflicting(
			path.MatchRoot("id"),
			path.MatchRoot("name"),
		),
	}
}

func (d *EdaCredentialTypeDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *EdaCredentialTypeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data EdaCredentialTypeModel

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
		url = fmt.Sprintf("credential-types/%d/", id)
	}
	if !data.Name.IsNull() {
		// set url for read by name HTTP request
		name := urlParser.QueryEscape(data.Name.ValueString())
		url = fmt.Sprintf("credential-types/?name=%s", name)
	}
	body, statusCode, err := d.client.GenericAPIRequest(ctx, http.MethodGet, url, nil, []int{200}, "eda")
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

	var responseData EdaCredentialTypeAPIModel

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
			Count   int                         `json:"count"`
			Results []EdaCredentialTypeAPIModel `json:"results"`
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

	data.Name = types.StringValue(responseData.Name)
	data.Kind = types.StringValue(responseData.Kind)

	if responseData.Description != "" {
		data.Description = types.StringValue(responseData.Description)
	}

	if responseData.Namespace != nil {
		namespaceStr, ok := responseData.Namespace.(string)
		if !ok {
			resp.Diagnostics.AddError("Unable to cast", "Unable to cast Namespace as string")
			return
		}
		data.Namespace = types.StringValue(namespaceStr)
	}

	rawInputs := responseData.Inputs
	rawInputsType := reflect.TypeOf(rawInputs)

	if responseData.Inputs != "" {
		if rawInputsType.Kind() == reflect.Map {
			rawInputs, ok := rawInputs.(map[string]any)
			if !ok {
				resp.Diagnostics.AddError("Unable to cast", "Unable to cast Inputs as map[string]any")
				return
			}

			if len(rawInputs) != 0 {
				tmpInputsMap := make(map[string]any, len(rawInputs))
				for k, v := range rawInputs {
					tmpInputsMap[k] = v
				}
				tmpInputsJson, err := json.Marshal(tmpInputsMap)
				if err != nil {
					resp.Diagnostics.AddError("Marshal issue", "Unable to marshal Inputs into json for storage.")
					return
				}
				data.Inputs = types.StringValue(string(tmpInputsJson))
			}
		}
	}
	if responseData.Injectors != "" {
		rawInjectors := responseData.Injectors
		rawInjectorsType := reflect.TypeOf(rawInjectors)

		if rawInjectorsType.Kind() == reflect.Map {
			rawInjectors, ok := rawInjectors.(map[string]any)
			if !ok {
				resp.Diagnostics.AddError("Unable to cast", "Unable to cast Injectors as map[string]any")
				return
			}

			if len(rawInjectors) != 0 {
				tmpInjectorsMap := make(map[string]any, len(rawInjectors))
				for k, v := range rawInjectors {
					tmpInjectorsMap[k] = v
				}
				tmpInjectorsJson, err := json.Marshal(tmpInjectorsMap)
				if err != nil {
					resp.Diagnostics.AddError("Marshal issue", "Unable to marshall Injectors into json for storage.")
					return
				}
				data.Injectors = types.StringValue(string(tmpInjectorsJson))
			}
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
