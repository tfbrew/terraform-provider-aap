package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &EdaCredentialResource{}
var _ resource.ResourceWithImportState = &EdaCredentialResource{}

func NewEdaCredentialResource() resource.Resource {
	return &EdaCredentialResource{}
}

type EdaCredentialResource struct {
	client *providerClient
}

func (r *EdaCredentialResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_eda_credential"
}

func (r *EdaCredentialResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `Manage an EDA credential. 
NOTE: The automation controller API does not return encrypted secrets so changes made in the controller of the inputs field will be ignored. 
The only changes to the inputs field that will be sent are when the terraform code does not match the terraform state.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Credential ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Credential name.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Credential description.",
				Optional:    true,
			},
			"organization_id": schema.Int32Attribute{
				Description: "ID of organization which owns this credential.",
				Optional:    true,
			},
			"credential_type_id": schema.Int32Attribute{
				Description: "ID of the credential type.",
				Required:    true,
			},
			"inputs": schema.DynamicAttribute{
				Description: "Specify a string by using using `jsonencode()` to encode similar data as as string in state. Specify alphabetically when using the second method.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

func (r *EdaCredentialResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *EdaCredentialResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data EdaCredentialModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var bodyData EdaCredentialAPIModel

	bodyData.Name = data.Name.ValueString()
	bodyData.CredentialTypeId = int(data.CredentialTypeId.ValueInt32())
	bodyData.OrganizationId = int(data.OrganizationId.ValueInt32())

	if !(data.Description.IsNull()) {
		bodyData.Description = data.Description.ValueString()
	}

	if !data.Inputs.IsUnderlyingValueNull() && !data.Inputs.IsNull() {
		inputsDataMap := make(map[string]any)

		switch val := data.Inputs.UnderlyingValue().(type) {
		case types.String:
			err := json.Unmarshal([]byte(val.ValueString()), &inputsDataMap)
			if err != nil {
				resp.Diagnostics.AddError(
					"Unable to unmarshal map to json",
					fmt.Sprintf("Unable to process inputs: %+v. ", data.Inputs))
				return
			}
		case types.Object:

			for key, v := range val.Attributes() {
				switch v := v.(type) {
				case types.String:
					// if the value is a string, we can use it as is
					inputsDataMap[key] = v.ValueString()
				case types.Bool:
					// if the value is a bool, we can use it as is
					inputsDataMap[key] = v.ValueBool()
				default:
					resp.Diagnostics.AddError(
						"inputs value specified is invalid type",
						fmt.Sprintf("inputs key '%s' has an unexpected type: %T", key, v),
					)
					return
				}
			}
		default:
			resp.Diagnostics.AddError("Inputs type invalid", "The inputs should be a types.String or types.Object.")
			return
		}

		bodyData.Inputs = inputsDataMap
	}

	url := "eda-credentials/"
	returnedData, _, err := r.client.CreateUpdateAPIRequest(ctx, http.MethodPost, url, bodyData, []int{201}, "eda")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API http request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}

	returnedValues := []string{"id"}
	for _, key := range returnedValues {
		if _, exists := returnedData[key]; !exists {
			resp.Diagnostics.AddError(
				"Error retrieving computed values",
				fmt.Sprintf("Could not retrieve %v.", key))
			return
		}
	}

	data.Id = types.StringValue(fmt.Sprintf("%v", returnedData["id"]))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EdaCredentialResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data EdaCredentialModel

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

	url := fmt.Sprintf("eda-credentials/%d/", id)
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

	var responseData EdaCredentialAPIModel

	err = json.Unmarshal(body, &responseData)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to unmarshal json",
			fmt.Sprintf("bodyData: %+v.", err))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), responseData.Name)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("credential_type_id"), responseData.CredentialType.Id)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), responseData.Organization.Id)...)

	if !data.Description.IsNull() || responseData.Description != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("description"), responseData.Description)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Handle inputs attribute.
	// This is dymanic and we document that they should provide a String or an Object for this attribute.
	// Inputs themselves will only be string or boolean, fyi: https://docs.redhat.com/en/documentation/red_hat_ansible_automation_platform/2.5/html/using_automation_decisions/eda-credential-types

	// we haven't imported it & not set in state previously
	if data.Inputs.IsUnderlyingValueNull() && responseData.Inputs != nil && len(responseData.Inputs) > 0 {
		resp.Diagnostics.Append(setInputfromResponeData(ctx, resp, &responseData.Inputs)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// we have imported something or state has values prevously; so,
	//    we need to try and get our value to match API regardless of order & $encrypted$ values

	if !data.Inputs.IsUnderlyingValueNull() && !data.Inputs.IsNull() {

		inputsValue := data.Inputs.UnderlyingValue()

		// convert state to map[string]any
		currInputsState := make(map[string]any)

		switch val := inputsValue.(type) {
		case types.Object:
			for key, v := range val.Attributes() {
				switch v := v.(type) {
				case types.String:
					// if the value is a string, we can use it as is
					currInputsState[key] = v.ValueString()
				case types.Bool:
					// if the value is a bool, we can use it as is
					currInputsState[key] = v.ValueBool()
				default:
					resp.Diagnostics.AddError(
						"inputs value specified is invalid type",
						fmt.Sprintf("inputs key '%s' has an unexpected type: %T", key, v),
					)
					return
				}
			}

			replaceEncryptedApiValues(&currInputsState, &responseData.Inputs)
			resp.Diagnostics.Append(setInputfromResponeData(ctx, resp, &responseData.Inputs)...)
			if resp.Diagnostics.HasError() {
				return
			}
		case types.String:
			if err := json.Unmarshal([]byte(val.ValueString()), &currInputsState); err != nil {
				resp.Diagnostics.AddError(
					"Unable to unmarshal inputs from string",
					fmt.Sprintf("Error: %v", err),
				)
				return
			}

			replaceEncryptedApiValues(&currInputsState, &responseData.Inputs)

			if !reflect.DeepEqual(currInputsState, responseData.Inputs) {
				// if they are not equal, we need to update state to match API - otherwise leave state as is
				inputsBytes, err := json.Marshal(responseData.Inputs)
				if err != nil {
					resp.Diagnostics.AddError(
						"Unable to marshal inputs to string",
						fmt.Sprintf("Error: %v", err),
					)
					return
				}
				stateInputs := types.DynamicValue(types.StringValue(string(inputsBytes)))
				resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("inputs"), stateInputs)...)
				if resp.Diagnostics.HasError() {
					return
				}
			}

		default:
			resp.Diagnostics.AddError("inputs value specified is invalid type", "inputs must be an object or string type.")
			return
		}
	}

}

func (r *EdaCredentialResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data EdaCredentialModel

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

	var bodyData EdaCredentialAPIModel

	bodyData.Name = data.Name.ValueString()
	bodyData.CredentialTypeId = int(data.CredentialTypeId.ValueInt32())
	bodyData.OrganizationId = int(data.OrganizationId.ValueInt32())

	if !(data.Description.IsNull()) {
		bodyData.Description = data.Description.ValueString()
	}

	if !data.Inputs.IsUnderlyingValueNull() {
		inputsDataMap := make(map[string]any)

		switch val := data.Inputs.UnderlyingValue().(type) {
		case types.String:
			err = json.Unmarshal([]byte(val.ValueString()), &inputsDataMap)
			if err != nil {
				resp.Diagnostics.AddError(
					"Unable to unmarshal map to json",
					fmt.Sprintf("Unable to process inputs: %+v. ", data.Inputs))
				return
			}
		case types.Object:
			for key, v := range val.Attributes() {
				switch v := v.(type) {
				case types.String:
					// if the value is a string, we can use it as is
					inputsDataMap[key] = v.ValueString()
				case types.Bool:
					// if the value is a bool, we can use it as is
					inputsDataMap[key] = v.ValueBool()
				default:
					resp.Diagnostics.AddError(
						"inputs value specified is invalid type",
						fmt.Sprintf("inputs key '%s' has an unexpected type: %T", key, v),
					)
					return
				}
			}
		default:
			resp.Diagnostics.AddError("Inputs type invalid", "The inputs should be a types.String or types.Object.")
			return
		}

		bodyData.Inputs = inputsDataMap
	}

	url := fmt.Sprintf("eda-credentials/%d/", id)
	returnedData, _, err := r.client.CreateUpdateAPIRequest(ctx, http.MethodPatch, url, bodyData, []int{200}, "eda")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API update request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}

	returnedValues := []string{"id"}
	for _, key := range returnedValues {
		if _, exists := returnedData[key]; !exists {
			resp.Diagnostics.AddError(
				"Error retrieving computed values",
				fmt.Sprintf("Could not retrieve %v.", key))
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EdaCredentialResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data EdaCredentialModel

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

	url := fmt.Sprintf("eda-credentials/%d/", id)
	_, _, err = r.client.GenericAPIRequest(ctx, http.MethodDelete, url, nil, []int{202, 204}, "eda")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API delete request",
			fmt.Sprintf("Error was: %s.", err.Error()))
		return
	}
}

func (r *EdaCredentialResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	idUnescaped, _ := strconv.Unquote(`"` + req.ID + `"`)

	idParts := strings.Split(idUnescaped, importIDSeparator)
	countParts := len(idParts)

	switch {
	case countParts == 1:
		resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	case ((countParts >= 3) && ((countParts-1)%2) == 0): // verify they provided pairs of values beyond the ID

		resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[0])...)
		if resp.Diagnostics.HasError() {
			return
		}

		inputsValues := make(map[string]attr.Value)
		inputsAttrTypes := make(map[string]attr.Type)

		for i := 1; i < countParts; i += 2 {
			inputsValues[idParts[i]] = types.StringValue(idParts[i+1])
			inputsAttrTypes[idParts[i]] = types.StringType
		}

		objVal, diag := types.ObjectValue(inputsAttrTypes, inputsValues)
		resp.Diagnostics.Append(diag...)
		if resp.Diagnostics.HasError() {
			return
		}
		dynamicVal := types.DynamicValue(objVal)

		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("inputs"), dynamicVal)...)
		if resp.Diagnostics.HasError() {
			return
		}

	default:
		resp.Diagnostics.AddError("Invalid import id string", "The import string at the end must contain one int id value or that plus comma-separated pairs for string keys with corresponding secrets.")

	}

}
