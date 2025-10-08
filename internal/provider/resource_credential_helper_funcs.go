package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func extraDataApiToDynamicObject(ctx context.Context, apiExtraData map[string]any, dynValue *basetypes.DynamicValue) diag.Diagnostics {
	extraData := apiExtraData
	extraDataValues := make(map[string]attr.Value)
	extraDataAttrTypes := make(map[string]attr.Type)

	for k, v := range extraData {
		switch val := v.(type) {
		case string:

			extraDataValues[k] = types.StringValue(val)
			extraDataAttrTypes[k] = types.StringType
		case []any:
			// Treat the slice as a string

			tfList, diags := types.ListValueFrom(ctx, types.StringType, val)
			if diags.HasError() {
				return diags
			}
			extraDataValues[k] = tfList
			extraDataAttrTypes[k] = types.ListType{ElemType: types.StringType}
		case float64:
			extraDataValues[k] = types.Float64Value(val)
			extraDataAttrTypes[k] = types.Float64Type

		default:
			diags := diag.Diagnostics{}
			diags.AddError(
				"Unexpected Extra Data Type",
				fmt.Sprintf("Extra Data '%s' has an unexpected type: %T", k, v),
			)
			return diags
		}
	}

	objVal, diag := types.ObjectValue(extraDataAttrTypes, extraDataValues)

	if diag.HasError() {
		return diag
	}

	*dynValue = types.DynamicValue(objVal)

	return nil
}

func credentialInputApiToDynamicObject(apiInputs map[string]any, dynValue *basetypes.DynamicValue) diag.Diagnostics {
	inputs := apiInputs
	inputsValues := make(map[string]attr.Value)
	inputsAttrTypes := make(map[string]attr.Type)

	for k, v := range inputs {
		switch val := v.(type) {
		case string:
			inputsValues[k] = types.StringValue(val)
			inputsAttrTypes[k] = types.StringType
		case bool:
			inputsValues[k] = types.BoolValue(val)
			inputsAttrTypes[k] = types.BoolType
		default:
			diags := diag.Diagnostics{}
			diags.AddError(
				"Unexpected Input Type",
				fmt.Sprintf("Input '%s' has an unexpected type: %T", k, v),
			)
			return diags
		}
	}

	objVal, diag := types.ObjectValue(inputsAttrTypes, inputsValues)

	if diag.HasError() {
		return diag
	}

	*dynValue = types.DynamicValue(objVal)

	return nil
}

func setInputfromResponeData(ctx context.Context, resp *resource.ReadResponse, responseData *CredentialAPIModel) diag.Diagnostics {
	var dynValue basetypes.DynamicValue
	diags := credentialInputApiToDynamicObject(responseData.Inputs, &dynValue)
	if diags.HasError() {
		return diags
	}

	diags.Append(resp.State.SetAttribute(ctx, path.Root("inputs"), &dynValue)...)
	return diags
}

func replaceEncryptedApiValues(currInputsState *map[string]any, responseData *CredentialAPIModel) {
	currInputsStateMap := *currInputsState
	// loop through API data and find/replace $encrypted$ values from state
	for k, v := range responseData.Inputs {
		switch val := v.(type) {
		case string:
			if val == "$encrypted$" {
				if currInputsStateMap[k] != nil && currInputsStateMap[k] != "" {
					responseData.Inputs[k] = currInputsStateMap[k]
				}
			}
		}
	}
}
