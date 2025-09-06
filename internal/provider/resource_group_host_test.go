package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/tfbrew/terraform-provider-aap/internal/configprefix"
)

func TestAccGroupHostResource(t *testing.T) {
	IdCompare := &compareTwoValuesAsStrings{}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGrpHstOrgInv() + testAccGrpHst1stPass() + testAccGrpHst1stPassGrp2(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_group_host.grp-host-link", configprefix.Prefix),
						tfjsonpath.New("group_id"),
						fmt.Sprintf("%s_group.group-example", configprefix.Prefix),
						tfjsonpath.New("id"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_group_host.grp-host-link", configprefix.Prefix),
						tfjsonpath.New("host_id"),
						fmt.Sprintf("%s_host.host-1", configprefix.Prefix),
						tfjsonpath.New("id"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_group_host.grp-host-link-2", configprefix.Prefix),
						tfjsonpath.New("group_id"),
						fmt.Sprintf("%s_group.group-example", configprefix.Prefix),
						tfjsonpath.New("id"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_group_host.grp-host-link-2", configprefix.Prefix),
						tfjsonpath.New("host_id"),
						fmt.Sprintf("%s_host.host-2", configprefix.Prefix),
						tfjsonpath.New("id"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_group_host.grp2-host-link", configprefix.Prefix),
						tfjsonpath.New("group_id"),
						fmt.Sprintf("%s_group.group-example-2", configprefix.Prefix),
						tfjsonpath.New("id"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_group_host.grp2-host-link", configprefix.Prefix),
						tfjsonpath.New("host_id"),
						fmt.Sprintf("%s_host.host-2", configprefix.Prefix),
						tfjsonpath.New("id"),
						IdCompare,
					),
				},
			},
			{
				Config: testAccGrpHstOrgInv() + testAccGrpHst1stPass() + testAccGrpHst2ndPassGrp2(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(fmt.Sprintf("%s_group_host.grp2-host-link", configprefix.Prefix), plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
			},
		},
	})
}

func testAccGrpHstOrgInv() string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "example" {
  name        = "%[2]s"
  description = "example"
}

resource "%[1]s_inventory" "example" {
  name         = "%[2]s"
  description  = "example"
  organization = %[1]s_organization.example.id
}	
	`, configprefix.Prefix, acctest.RandString(5))
}

func testAccGrpHst1stPass() string {
	return fmt.Sprintf(`
resource "%[1]s_group" "group-example" {
  name        = "%[2]s"
  description = "Example with jsonencoded variables."
  inventory   = %[1]s_inventory.example.id
  variables = jsonencode(
    {
      foo = "bar"
      baz = "qux"
    }
  )
}

resource "%[1]s_host" "host-1" {
  name = "%[2]s-1"
  inventory = %[1]s_inventory.example.id
}

resource "%[1]s_host" "host-2" {
  name = "%[2]s-2"
  inventory = %[1]s_inventory.example.id
}

resource "%[1]s_group_host" "grp-host-link" {
  group_id = %[1]s_group.group-example.id
  host_id = %[1]s_host.host-1.id
}

resource "%[1]s_group_host" "grp-host-link-2" {
  group_id = %[1]s_group.group-example.id
  host_id = %[1]s_host.host-2.id
}

	`, configprefix.Prefix, acctest.RandString(5))
}

func testAccGrpHst1stPassGrp2() string {
	return fmt.Sprintf(`
resource "%[1]s_group" "group-example-2" {
  name        = "%[2]s"
  description = "A second group example."
  inventory   = %[1]s_inventory.example.id
}

resource "%[1]s_group_host" "grp2-host-link" {
  group_id = %[1]s_group.group-example-2.id
  host_id = %[1]s_host.host-2.id
}	
	`, configprefix.Prefix, acctest.RandString(5))
}

func testAccGrpHst2ndPassGrp2() string {
	return fmt.Sprintf(`
resource "%[1]s_group" "group-example-2" {
  name        = "%[2]s"
  description = "A second group example."
  inventory   = %[1]s_inventory.example.id
}

resource "%[1]s_group_host" "grp2-host-link" {
  group_id = %[1]s_group.group-example-2.id
  host_id = %[1]s_host.host-1.id
}	
	`, configprefix.Prefix, acctest.RandString(5))
}
