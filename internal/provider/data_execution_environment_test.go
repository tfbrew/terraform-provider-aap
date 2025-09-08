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

func TestAccExecutionEnvironmentDataSource(t *testing.T) {
	resource1 := ExecutionEnvironmentAPIModel{
		Name:        "test-ee-" + acctest.RandString(5),
		Description: "test execution environment",
		Image:       "quay.io/ansible/awx-ee:latest",
		Pull:        "always",
	}
	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0), // built-in check from tfversion package
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read by ID testing
			{
				Config: testAccExecutionEnvironmentDataSourceConfig(resource1),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_execution_environment.test-id", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource1.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_execution_environment.test-id", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(resource1.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_execution_environment.test-id", configprefix.Prefix),
						tfjsonpath.New("image"),
						knownvalue.StringExact(resource1.Image),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_execution_environment.test-id", configprefix.Prefix),
						tfjsonpath.New("pull"),
						knownvalue.StringExact(resource1.Pull),
					),
				},
			},
			// Read by name testing
			{
				Config: testAccExecutionEnvironmentDataSourceConfig(resource1),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_execution_environment.test-name", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource1.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_execution_environment.test-name", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(resource1.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_execution_environment.test-name", configprefix.Prefix),
						tfjsonpath.New("image"),
						knownvalue.StringExact(resource1.Image),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_execution_environment.test-name", configprefix.Prefix),
						tfjsonpath.New("pull"),
						knownvalue.StringExact(resource1.Pull),
					),
				},
			},
		},
	})
}

func testAccExecutionEnvironmentDataSourceConfig(resource ExecutionEnvironmentAPIModel) string {
	return fmt.Sprintf(`
resource "%[1]s_execution_environment" "test" {
  name        	= "%[2]s"
  description 	= "%[3]s"
  image   		= "%[4]s"
  pull 			= "%[5]s"
}
data "%[1]s_execution_environment" "test-id" {
  id = %[1]s_execution_environment.test.id
}
data "%[1]s_execution_environment" "test-name" {
  name = %[1]s_execution_environment.test.name
}
`, configprefix.Prefix, resource.Name, resource.Description, resource.Image, resource.Pull)
}
