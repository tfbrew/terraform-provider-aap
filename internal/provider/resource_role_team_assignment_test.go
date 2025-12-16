package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/tfbrew/terraform-provider-aap/internal/configprefix"
)

func TestAccRoleTeamAssignmentResource(t *testing.T) {
	IdCompare := &compareTwoValuesAsStrings{}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleTeamAssignmentResourceConfig(1),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_role_team_assignment.test", configprefix.Prefix),
						tfjsonpath.New("object_id"),
						fmt.Sprintf("%s_organization.test-1", configprefix.Prefix),
						tfjsonpath.New("aap25_gateway_id"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_role_team_assignment.test", configprefix.Prefix),
						tfjsonpath.New("role_definition"),
						fmt.Sprintf("%s_role_definition.test-1", configprefix.Prefix),
						tfjsonpath.New("id"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_role_team_assignment.test", configprefix.Prefix),
						tfjsonpath.New("team"),
						fmt.Sprintf("%s_team.test-1", configprefix.Prefix),
						tfjsonpath.New("id"),
						IdCompare,
					),
				},
			},
			{
				ResourceName:      fmt.Sprintf("%s_role_team_assignment.test", configprefix.Prefix),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRoleTeamAssignmentResourceConfig(2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_role_team_assignment.test", configprefix.Prefix),
						tfjsonpath.New("object_id"),
						fmt.Sprintf("%s_organization.test-2", configprefix.Prefix),
						tfjsonpath.New("aap25_gateway_id"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_role_team_assignment.test", configprefix.Prefix),
						tfjsonpath.New("role_definition"),
						fmt.Sprintf("%s_role_definition.test-2", configprefix.Prefix),
						tfjsonpath.New("id"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_role_team_assignment.test", configprefix.Prefix),
						tfjsonpath.New("team"),
						fmt.Sprintf("%s_team.test-2", configprefix.Prefix),
						tfjsonpath.New("id"),
						IdCompare,
					),
				},
			},
		},
	})
}

func testAccRoleTeamAssignmentResourceConfig(number int) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "test-%[3]d" {
  name        			= "%[2]s-%[3]d"
}
resource "%[1]s_role_definition" "test-%[3]d" {
  name         = "%[2]s"
  description  = "Test role definition"
  content_type = "shared.organization"
  permissions   = ["shared.view_organization"]
}
resource "%[1]s_team" "test-%[3]d" {
  name      = "%[2]s-%[3]d"
  organization   = %[1]s_organization.test-%[3]d.aap25_gateway_id
  description  = "%[2]s-%[3]d description"
}
resource "%[1]s_role_team_assignment" "test" {
  object_id       = %[1]s_organization.test-%[3]d.aap25_gateway_id
  role_definition = %[1]s_role_definition.test-%[3]d.id
  team            = %[1]s_team.test-%[3]d.id
}
`, configprefix.Prefix, acctest.RandString(5), number)
}
