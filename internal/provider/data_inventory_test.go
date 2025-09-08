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

func TestAccInventoryDataSource(t *testing.T) {
	IdCompare := &compareTwoValuesAsStrings{}
	resource1 := InventoryAPIModel{
		Name:         "test-inventory-" + acctest.RandString(5),
		Description:  "test description 1",
		Organization: 1,
		Variables:    "{\"foo\":\"bar\"}",
	}
	resource2 := InventoryAPIModel{
		Name:         "test-inventory-" + acctest.RandString(5),
		Description:  "test description 3",
		Organization: 1,
		Variables:    "{\"foo\":\"baz\"}",
		Kind:         "smart",
		HostFilter:   "name__icontains=localhost",
	}
	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0), // built-in check from tfversion package
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read normal inventory by ID
			{
				Config: testAccInventoryDataSource1Config(resource1),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_inventory.test1", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource1.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_inventory.test1", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(resource1.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_inventory.test1", configprefix.Prefix),
						tfjsonpath.New("variables"),
						knownvalue.StringExact(resource1.Variables),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_inventory.test1", configprefix.Prefix),
						tfjsonpath.New("kind"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_inventory.test1", configprefix.Prefix),
						tfjsonpath.New("host_filter"),
						knownvalue.Null(),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_organization.test1", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("data.%s_inventory.test1", configprefix.Prefix),
						tfjsonpath.New("organization"),
						IdCompare,
					),
				},
			},
			// Read smart inventory by ID
			{
				Config: testAccInventoryDataSource2Config(resource2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_inventory.test2", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource2.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_inventory.test2", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(resource2.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_inventory.test2", configprefix.Prefix),
						tfjsonpath.New("variables"),
						knownvalue.StringExact(resource2.Variables),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_inventory.test2", configprefix.Prefix),
						tfjsonpath.New("kind"),
						knownvalue.StringExact(resource2.Kind),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_inventory.test2", configprefix.Prefix),
						tfjsonpath.New("host_filter"),
						knownvalue.StringExact(resource2.HostFilter),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_organization.test2", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("data.%s_inventory.test2", configprefix.Prefix),
						tfjsonpath.New("organization"),
						IdCompare,
					),
				},
			},
		},
	})
}

func testAccInventoryDataSource1Config(resource InventoryAPIModel) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "test1" {
  name        			= "%[2]s"
}
resource "%[1]s_inventory" "test1" {
  name         = "%[3]s"
  description  = "%[4]s"
  organization = %[1]s_organization.test1.id
  variables    = jsonencode(%[5]s)
}
data "%[1]s_inventory" "test1" {
  id = %[1]s_inventory.test1.id
}
`, configprefix.Prefix, acctest.RandString(5), resource.Name, resource.Description, resource.Variables)
}

func testAccInventoryDataSource2Config(resource InventoryAPIModel) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "test2" {
  name        			= "%[2]s"
}
resource "%[1]s_inventory" "test2" {
  name         	= "%[3]s"
  description  	= "%[4]s"
  organization  = %[1]s_organization.test2.id
  variables    	= jsonencode(%[5]s)
  kind			= "%[6]s"
  host_filter	= "%[7]s"
}
data "%[1]s_inventory" "test2" {
  id = %[1]s_inventory.test2.id
}
`, configprefix.Prefix, acctest.RandString(5), resource.Name, resource.Description, resource.Variables, resource.Kind, resource.HostFilter)
}
