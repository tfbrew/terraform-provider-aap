package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/tfbrew/terraform-provider-aap/internal/configprefix"
)

func TestAccGroupResource(t *testing.T) {
	IdCompare := &compareTwoValuesAsStrings{}

	Group1 := GroupAPIModel{
		Name:        "test-group-" + acctest.RandString(5),
		Description: "Example 1",
		Variables:   "{\"foo\":\"bar\"}",
	}

	Group2 := GroupAPIModel{
		Name:        "test-group-" + acctest.RandString(5),
		Description: "Example 2",
		Variables:   "{\"baz\":\"qux\"}",
	}

	ReplacementInventory := InventoryAPIModel{
		Name:        "test-inventory-" + acctest.RandString(5),
		Description: "Example Replacement",
	}
	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupResourceConfig(Group1),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_group.test", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(Group1.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_group.test", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(Group1.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_group.test", configprefix.Prefix),
						tfjsonpath.New("variables"),
						knownvalue.StringExact(Group1.Variables),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_inventory.test", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_group.test", configprefix.Prefix),
						tfjsonpath.New("inventory"),
						IdCompare,
					),
				},
			},
			{
				ResourceName:      fmt.Sprintf("%s_group.test", configprefix.Prefix),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGroupResource2Config(Group2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_group.test2", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(Group2.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_group.test2", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(Group2.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_group.test2", configprefix.Prefix),
						tfjsonpath.New("variables"),
						knownvalue.StringExact(Group2.Variables),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_inventory.test2", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_group.test2", configprefix.Prefix),
						tfjsonpath.New("inventory"),
						IdCompare,
					),
				},
			},
			// change an existing inventory on an existing group and verify plan marks for re-create
			{
				Config: testAccSecondInvResourceConfig(ReplacementInventory) + testAccInvPlanRecreateConfig(Group2),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(fmt.Sprintf("%s_group.test2", configprefix.Prefix), plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
			},
		},
	})
}

func testAccSecondInvResourceConfig(resource InventoryAPIModel) string {
	return fmt.Sprintf(`
	resource "%[1]s_organization" "new-inv-test-org" {
		name = "test-organizatio-%[2]s"
	}
	resource "%[1]s_inventory" "new-inventory" {
		name = "%[3]s"
		organization = %[1]s_organization.new-inv-test-org.id
	}
		`, configprefix.Prefix, acctest.RandString(5), resource.Name)
}

// touches same group created by testAccGroupResource2Config(), but just changes the inventory value,
//
//	doing this to verify that the replace functionality works when updating this.
func testAccInvPlanRecreateConfig(resource GroupAPIModel) string {
	return fmt.Sprintf(`
resource "%[1]s_group" "test2" {
  name        = "%[2]s"
  description = "%[3]s"
  inventory   = %[1]s_inventory.new-inventory.id
  variables   = jsonencode(%[4]s)
}
  `, configprefix.Prefix, resource.Name, resource.Description, resource.Variables)
}

func testAccGroupResourceConfig(resource GroupAPIModel) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "test" {
  name        = "test-organization-%[2]s"
  description = "test"
}
resource "%[1]s_inventory" "test" {
  name         = "test-inventory-%[2]s"
  description  = "test"
  organization = %[1]s_organization.test.id
}
resource "%[1]s_group" "test" {
  name        = "%[3]s"
  description = "%[4]s"
  inventory   = %[1]s_inventory.test.id
  variables   = jsonencode(%[5]s)
}
  `, configprefix.Prefix, acctest.RandString(5), resource.Name, resource.Description, resource.Variables)
}

func testAccGroupResource2Config(resource GroupAPIModel) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "test2" {
  name        = "test-organization-%[2]s"
  description = "test"
}
resource "%[1]s_inventory" "test2" {
  name         = "test-inventory-%[2]s"
  description  = "test"
  organization = %[1]s_organization.test2.id
}
resource "%[1]s_group" "test2" {
  name        = "%[3]s"
  description = "%[4]s"
  inventory   = %[1]s_inventory.test2.id
  variables   = jsonencode(%[5]s)
}
  `, configprefix.Prefix, acctest.RandString(5), resource.Name, resource.Description, resource.Variables)
}
