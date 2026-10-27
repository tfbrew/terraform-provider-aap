// AAP-AWX-DIFF-FILE: AAP-only. Event-Driven Ansible (EDA) has no AWX/AAP 2.4 equivalent.
package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type EdaCredentialModel struct {
	Id               types.String  `tfsdk:"id"`
	Name             types.String  `tfsdk:"name"`
	Description      types.String  `tfsdk:"description"`
	OrganizationId   types.Int32   `tfsdk:"organization_id"`
	CredentialTypeId types.Int32   `tfsdk:"credential_type_id"`
	Inputs           types.Dynamic `tfsdk:"inputs"`
}

type EdaCredentialAPIModel struct {
	Id               int                        `json:"id"`
	Name             string                     `json:"name"`
	Description      string                     `json:"description,omitempty"`
	OrganizationId   int                        `json:"organization_id"`
	Organization     OrganizationNestedAPIModel `json:"organization,omitempty"`
	CredentialTypeId int                        `json:"credential_type_id"`
	CredentialType   CredentialNestedAPIModel   `json:"credential_type,omitempty"`
	Inputs           map[string]any             `json:"inputs"`
}

type EdaCredentialDataModel struct {
	Id               types.String  `tfsdk:"id"`
	Name             types.String  `tfsdk:"name"`
	Description      types.String  `tfsdk:"description"`
	OrganizationId   types.Int32   `tfsdk:"organization_id"`
	CredentialTypeId types.Int32   `tfsdk:"credential_type_id"`
	Inputs           types.String  `tfsdk:"inputs"`
	InputsAsObject   types.Dynamic `tfsdk:"inputs_as_object"`
}

type CredentialNestedAPIModel struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type EdaCredentialTypeModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Inputs      types.String `tfsdk:"inputs"`
	Injectors   types.String `tfsdk:"injectors"`
	Kind        types.String `tfsdk:"kind"`
	Namespace   types.String `tfsdk:"namespace"`
}

type EdaCredentialTypeAPIModel struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Inputs      any    `json:"inputs"`
	Injectors   any    `json:"injectors"`
	Kind        string `json:"kind"`
	Namespace   any    `json:"namespace"`
}

type EdaDecisionEnvironmentModel struct {
	Id              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	EdaCredentialId types.Int32  `tfsdk:"eda_credential_id"`
	OrganizationId  types.Int32  `tfsdk:"organization_id"`
	ImageUrl        types.String `tfsdk:"image_url"`
	PullPolicy      types.String `tfsdk:"pull_policy"`
}

type EdaDecisionEnvironmentAPIModel struct {
	Id              int                        `json:"id"`
	Name            string                     `json:"name"`
	Description     string                     `json:"description,omitempty"`
	EdaCredential   CredentialNestedAPIModel   `json:"eda_credential,omitempty"`
	EdaCredentialId int                        `json:"eda_credential_id,omitempty"`
	Organization    OrganizationNestedAPIModel `json:"organization,omitempty"`
	OrganizationId  int                        `json:"organization_id"`
	ImageUrl        string                     `json:"image_url"`
	PullPolicy      string                     `json:"pull_policy"`
}

type EdaProjectModel struct {
	Id                              types.String `tfsdk:"id"`
	Name                            types.String `tfsdk:"name"`
	Description                     types.String `tfsdk:"description"`
	EdaCredentialId                 types.Int32  `tfsdk:"eda_credential_id"`
	OrganizationId                  types.Int32  `tfsdk:"organization_id"`
	Proxy                           types.String `tfsdk:"proxy"`
	ScmBranch                       types.String `tfsdk:"scm_branch"`
	ScmRefSpec                      types.String `tfsdk:"scm_refspec"`
	ScmType                         types.String `tfsdk:"scm_type"`
	SignatureValidationCredentialId types.Int32  `tfsdk:"signature_validation_credential_id"`
	Url                             types.String `tfsdk:"url"`
	VerifySsl                       types.Bool   `tfsdk:"verify_ssl"`
}

type EdaProjectAPIModel struct {
	Id                              int                        `json:"id"`
	Name                            string                     `json:"name"`
	Description                     string                     `json:"description,omitempty"`
	EdaCredentialId                 int                        `json:"eda_credential_id,omitempty"`
	Organization                    OrganizationNestedAPIModel `json:"organization"`
	OrganizationId                  int                        `json:"organization_id"`
	Proxy                           string                     `json:"proxy,omitempty"`
	ScmBranch                       string                     `json:"scm_branch,omitempty"`
	ScmRefSpec                      string                     `json:"scm_refspec,omitempty"`
	ScmType                         string                     `json:"scm_type"`
	SignatureValidationCredentialId int                        `json:"signature_validation_credential_id,omitempty"`
	Url                             string                     `json:"url"`
	VerifySsl                       bool                       `json:"verify_ssl"`
}

type OrganizationNestedAPIModel struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
