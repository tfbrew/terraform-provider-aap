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

func TestAccHostDataSource(t *testing.T) {
	host := HostAPIModel{
		Name:        "test-host-" + acctest.RandString(5),
		Description: "Example with jsonencoded variables for localhost",
		Variables:   "{\"foo\":\"bar\"}",
		Enabled:     true,
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
				Config: testAccHostDataSourceConfig(host),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_host.test", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(host.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_host.test", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(host.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_host.test", configprefix.Prefix),
						tfjsonpath.New("variables"),
						knownvalue.StringExact(host.Variables),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_host.test", configprefix.Prefix),
						tfjsonpath.New("enabled"),
						knownvalue.Bool(host.Enabled),
					),
				},
			},
		},
	})
}

func testAccHostDataSourceConfig(resource HostAPIModel) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "example" {
  name        = "test-organization-%[2]s"
  description = "test"
}
resource "%[1]s_inventory" "example" {
  name         = "test-inventory-%[2]s"
  description  = "test"
  organization = %[1]s_organization.example.id
}
resource "%[1]s_host" "test" {
  name        = "%[3]s"
  description = "%[4]s"
  inventory   = %[1]s_inventory.example.id
  variables   = jsonencode(%[5]s)
  enabled 	  = %[6]v
}
data "%[1]s_host" "test" {
  id = %[1]s_host.test.id
}
`, configprefix.Prefix, acctest.RandString(5), resource.Name, resource.Description, resource.Variables, resource.Enabled)
}
