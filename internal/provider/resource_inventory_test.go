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

func TestAccInventoryResource(t *testing.T) {
	rName := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	rName2 := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	rName3 := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	IdCompare := &compareTwoValuesAsStrings{}
	resource1 := InventoryAPIModel{
		Name:        "test-inventory-" + acctest.RandString(5),
		Description: "test description 1",
		Variables:   "{\"foo\":\"bar\"}",
	}
	resource2 := InventoryAPIModel{
		Name:        "test-inventory-" + acctest.RandString(5),
		Description: "test description 2",
		Variables:   "{\"foo\":\"baz\"}",
	}
	resource3 := InventoryAPIModel{
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
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInventoryResource1Config(resource1, rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_inventory.%s", configprefix.Prefix, rName),
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource1.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_inventory.%s", configprefix.Prefix, rName),
						tfjsonpath.New("description"),
						knownvalue.StringExact(resource1.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_inventory.%s", configprefix.Prefix, rName),
						tfjsonpath.New("variables"),
						knownvalue.StringExact(resource1.Variables),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_inventory.%s", configprefix.Prefix, rName),
						tfjsonpath.New("kind"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_inventory.%s", configprefix.Prefix, rName),
						tfjsonpath.New("host_filter"),
						knownvalue.Null(),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_organization.%s", configprefix.Prefix, rName),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_inventory.%s", configprefix.Prefix, rName),
						tfjsonpath.New("organization"),
						IdCompare,
					),
				},
			},
			{
				ResourceName:      fmt.Sprintf("%s_inventory.%s", configprefix.Prefix, rName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInventoryResource1Config(resource2, rName2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_inventory.%s", configprefix.Prefix, rName2),
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource2.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_inventory.%s", configprefix.Prefix, rName2),
						tfjsonpath.New("description"),
						knownvalue.StringExact(resource2.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_inventory.%s", configprefix.Prefix, rName2),
						tfjsonpath.New("variables"),
						knownvalue.StringExact(resource2.Variables),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_inventory.%s", configprefix.Prefix, rName2),
						tfjsonpath.New("kind"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_inventory.%s", configprefix.Prefix, rName2),
						tfjsonpath.New("host_filter"),
						knownvalue.Null(),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_organization.%s", configprefix.Prefix, rName2),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_inventory.%s", configprefix.Prefix, rName2),
						tfjsonpath.New("organization"),
						IdCompare,
					),
				},
			},
			{
				Config: testAccInventoryResource3Config(resource3, rName3),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_inventory.%s", configprefix.Prefix, rName3),
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource3.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_inventory.%s", configprefix.Prefix, rName3),
						tfjsonpath.New("description"),
						knownvalue.StringExact(resource3.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_inventory.%s", configprefix.Prefix, rName3),
						tfjsonpath.New("variables"),
						knownvalue.StringExact(resource3.Variables),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_inventory.%s", configprefix.Prefix, rName3),
						tfjsonpath.New("kind"),
						knownvalue.StringExact(resource3.Kind),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_inventory.%s", configprefix.Prefix, rName3),
						tfjsonpath.New("host_filter"),
						knownvalue.StringExact(resource3.HostFilter),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_organization.%s", configprefix.Prefix, rName3),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_inventory.%s", configprefix.Prefix, rName3),
						tfjsonpath.New("organization"),
						IdCompare,
					),
				},
			},
		},
	})
}

func testAccInventoryResource1Config(resource InventoryAPIModel, rName string) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "%[6]s" {
  name        			= "%[2]s"
}
resource "%[1]s_inventory" "%[6]s" {
  name         = "%[3]s"
  description  = "%[4]s"
  organization = %[1]s_organization.%[6]s.id
  variables    = jsonencode(%[5]s)
}
  `, configprefix.Prefix, acctest.RandString(5), resource.Name, resource.Description, resource.Variables, rName)
}

func testAccInventoryResource3Config(resource InventoryAPIModel, rName string) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "%[8]s" {
  name        			= "%[2]s"
}
resource "%[1]s_inventory" "%[8]s" {
  name         	= "%[3]s"
  description  	= "%[4]s"
  organization 	= %[1]s_organization.%[8]s.id
  variables    	= jsonencode(%[5]s)
  kind			= "%[6]s"
  host_filter	= "%[7]s"
}
  `, configprefix.Prefix, acctest.RandString(5), resource.Name, resource.Description, resource.Variables, resource.Kind, resource.HostFilter, rName)
}
