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

func TestAccGroupDataSource(t *testing.T) {
	group := GroupAPIModel{
		Name:        "test-group-" + acctest.RandString(5),
		Description: "Example group for datasource test",
		Variables:   "{\"foo\":\"bar\"}",
		Inventory:   0, // will be set in config
	}
	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupDataSourceConfig(group),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_group.test", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(group.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_group.test", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(group.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_group.test", configprefix.Prefix),
						tfjsonpath.New("variables"),
						knownvalue.StringExact(group.Variables),
					),
				},
			},
			// Lookup by name and inventory
			{
				Config: testAccGroupDataSourceConfigByName(group),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_group.by_name", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(group.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_group.by_name", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(group.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_group.by_name", configprefix.Prefix),
						tfjsonpath.New("variables"),
						knownvalue.StringExact(group.Variables),
					),
				},
			},
		},
	})
}

func testAccGroupDataSourceConfig(resource GroupAPIModel) string {
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
resource "%[1]s_group" "test" {
  name        = "%[3]s"
  description = "%[4]s"
  inventory   = %[1]s_inventory.example.id
  variables   = jsonencode(%[5]s)
}
data "%[1]s_group" "test" {
  id = %[1]s_group.test.id
}
`, configprefix.Prefix, acctest.RandString(5), resource.Name, resource.Description, resource.Variables)
}

func testAccGroupDataSourceConfigByName(resource GroupAPIModel) string {
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
resource "%[1]s_group" "test" {
  name        = "%[3]s"
  description = "%[4]s"
  inventory   = %[1]s_inventory.example.id
  variables   = jsonencode(%[5]s)
}
data "%[1]s_group" "by_name" {
  name      = %[1]s_group.test.name
  inventory = %[1]s_inventory.example.id
}
`, configprefix.Prefix, acctest.RandString(5), resource.Name, resource.Description, resource.Variables)
}
