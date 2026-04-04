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

func TestAccExecutionEnvironmentResource(t *testing.T) {
	rName := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	r2Name := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	IdCompare := &compareTwoValuesAsStrings{}
	resource1 := ExecutionEnvironmentAPIModel{
		Name:        "test-ee-" + acctest.RandString(5),
		Description: "test execution environment1",
		Image:       "quay.io/ansible/awx-ee:latest",
		Pull:        "always",
	}

	resource2 := ExecutionEnvironmentAPIModel{
		Name:        "test-ee-" + acctest.RandString(5),
		Description: "test execution environment2",
		Image:       "quay.io/ansible/awx-ee:latest",
		Pull:        "never",
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccExecutionEnvironmentResource1Config(resource1, rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_execution_environment.%s", configprefix.Prefix, rName),
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource1.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_execution_environment.%s", configprefix.Prefix, rName),
						tfjsonpath.New("description"),
						knownvalue.StringExact(resource1.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_execution_environment.%s", configprefix.Prefix, rName),
						tfjsonpath.New("image"),
						knownvalue.StringExact(resource1.Image),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_execution_environment.%s", configprefix.Prefix, rName),
						tfjsonpath.New("pull"),
						knownvalue.StringExact(resource1.Pull),
					),
				},
			},
			{
				ResourceName:      fmt.Sprintf("%s_execution_environment.%s", configprefix.Prefix, rName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccExecutionEnvironmentResource2Config(resource2, r2Name),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_execution_environment.%s", configprefix.Prefix, r2Name),
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource2.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_execution_environment.%s", configprefix.Prefix, r2Name),
						tfjsonpath.New("description"),
						knownvalue.StringExact(resource2.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_execution_environment.%s", configprefix.Prefix, r2Name),
						tfjsonpath.New("image"),
						knownvalue.StringExact(resource2.Image),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_execution_environment.%s", configprefix.Prefix, r2Name),
						tfjsonpath.New("pull"),
						knownvalue.StringExact(resource2.Pull),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_credential.%s", configprefix.Prefix, r2Name),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_execution_environment.%s", configprefix.Prefix, r2Name),
						tfjsonpath.New("credential"),
						IdCompare,
					),
				},
			},
		},
	})
}

func testAccExecutionEnvironmentResource1Config(resource ExecutionEnvironmentAPIModel, rName string) string {
	return fmt.Sprintf(`
resource "%[1]s_execution_environment" "%[6]s" {
  name        	= "%[2]s"
  description 	= "%[3]s"
  image   		= "%[4]s"
  pull 			= "%[5]s"
}
  `, configprefix.Prefix, resource.Name, resource.Description, resource.Image, resource.Pull, rName)
}

func testAccExecutionEnvironmentResource2Config(resource ExecutionEnvironmentAPIModel, rName string) string {
	return fmt.Sprintf(`
data "%[1]s_credential_type" "%[7]s" {
  name = "Container Registry"
  kind = "registry"
}
resource "%[1]s_organization" "%[7]s" {
  name        = "%[2]s"
}
resource "%[1]s_credential" "%[7]s" {
  name            = "%[2]s"
  organization    = %[1]s_organization.%[7]s.id
  credential_type = data.%[1]s_credential_type.%[7]s.id
  inputs = jsonencode({
	"host" : "quay.io",
	"username" : "test",
	"password" : "%[2]s",
	"verify_ssl" : true
  })
}
resource "%[1]s_execution_environment" "%[7]s" {
  name        	= "%[3]s"
  description 	= "%[4]s"
  image   		= "%[5]s"
  pull 			= "%[6]s"
  credential	= %[1]s_credential.%[7]s.id
}
  `, configprefix.Prefix, acctest.RandString(5), resource.Name, resource.Description, resource.Image, resource.Pull, rName)
}
