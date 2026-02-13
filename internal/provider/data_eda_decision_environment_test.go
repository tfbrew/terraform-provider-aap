package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/tfbrew/terraform-provider-aap/internal/configprefix"
)

func TestAccEdaDecisionEnvironmentDataSource(t *testing.T) {
	IdCompare := &compareTwoValuesAsStrings{}
	edadecisionenvironment1 := EdaDecisionEnvironmentAPIModel{
		Name:        "test-decision-environment-" + acctest.RandString(5),
		Description: "Test 1",
		ImageUrl:    "quay.io/ansible/awx-latest",
		PullPolicy:  "always",
	}

	edadecisionenvironment2 := EdaDecisionEnvironmentAPIModel{
		Name:        "test-decision-environment-" + acctest.RandString(5),
		Description: "Test 2",
		ImageUrl:    "repo/decisionenvironment/image-name:tag",
		PullPolicy:  "never",
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEdaDecisionEnvironmentDataSourceConfig(edadecisionenvironment1),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_eda_decision_environment.test", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(edadecisionenvironment1.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_eda_decision_environment.test", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(edadecisionenvironment1.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_eda_decision_environment.test", configprefix.Prefix),
						tfjsonpath.New("image_url"),
						knownvalue.StringExact(edadecisionenvironment1.ImageUrl),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_eda_decision_environment.test", configprefix.Prefix),
						tfjsonpath.New("pull_policy"),
						knownvalue.StringExact(edadecisionenvironment1.PullPolicy),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_organization.test", configprefix.Prefix),
						tfjsonpath.New("eda_id"),
						fmt.Sprintf("data.%s_eda_decision_environment.test", configprefix.Prefix),
						tfjsonpath.New("organization_id"),
						IdCompare,
					),
				},
			},
			{
				Config: testAccEdaDecisionEnvironmentDataSourceConfig2(edadecisionenvironment2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_eda_decision_environment.test-container-registry", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(edadecisionenvironment2.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_eda_decision_environment.test-container-registry", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(edadecisionenvironment2.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_eda_decision_environment.test-container-registry", configprefix.Prefix),
						tfjsonpath.New("image_url"),
						knownvalue.StringExact(edadecisionenvironment2.ImageUrl),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_eda_decision_environment.test-container-registry", configprefix.Prefix),
						tfjsonpath.New("pull_policy"),
						knownvalue.StringExact(edadecisionenvironment2.PullPolicy),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_organization.test-container-registry", configprefix.Prefix),
						tfjsonpath.New("eda_id"),
						fmt.Sprintf("data.%s_eda_decision_environment.test-container-registry", configprefix.Prefix),
						tfjsonpath.New("organization_id"),
						IdCompare,
					),
				},
			},
		},
	})
}

func testAccEdaDecisionEnvironmentDataSourceConfig(resource EdaDecisionEnvironmentAPIModel) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "test" {
  name        			= "%[2]s"
}
resource "%[1]s_eda_decision_environment" "test" {
  name         			= "%[3]s"
  description  			= "%[4]s"
  image_url      	    = "%[5]s"
  organization_id    	= %[1]s_organization.test.eda_id
  pull_policy        	= "%[6]s"
}
data "%[1]s_eda_decision_environment" "test" {
  id = %[1]s_eda_decision_environment.test.id
}
  `, configprefix.Prefix, acctest.RandString(5), resource.Name, resource.Description, resource.ImageUrl, resource.PullPolicy)
}

func testAccEdaDecisionEnvironmentDataSourceConfig2(resource EdaDecisionEnvironmentAPIModel) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "test-container-registry" {
  name        			= "%[2]s"
}
data "%[1]s_eda_credential_type" "test-container-registry" {
  name = "Container Registry"
}
resource "%[1]s_eda_credential" "test-container-registry" {
  name            = "%[3]s"
  description	  = "%[4]s"
  organization_id = %[1]s_organization.test-container-registry.eda_id
  credential_type_id = data.%[1]s_eda_credential_type.test-container-registry.id
  inputs = jsonencode({
			"host":       "quay.io",
			"password":   "test1234",
			"username":   "test",
			"verify_ssl": true,
		})
}
resource "%[1]s_eda_decision_environment" "test-container-registry" {
  name         			= "%[3]s"
  description  			= "%[4]s"
  image_url      	    = "%[5]s"
  organization_id    	= %[1]s_organization.test-container-registry.eda_id
  pull_policy        	= "%[6]s"
}
data "%[1]s_eda_decision_environment" "test-container-registry" {
  name = %[1]s_eda_decision_environment.test-container-registry.name
}
  `, configprefix.Prefix, acctest.RandString(5), resource.Name, resource.Description, resource.ImageUrl, resource.PullPolicy)
}
