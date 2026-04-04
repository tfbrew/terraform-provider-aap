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
	rName := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)

	IdCompare := &compareTwoValuesAsStrings{}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleTeamAssignmentResourceConfig(1, rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_role_team_assignment.%s", configprefix.Prefix, rName),
						tfjsonpath.New("object_id"),
						fmt.Sprintf("%s_organization.%s", configprefix.Prefix, rName+"-1"),
						tfjsonpath.New("id"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_role_team_assignment.%s", configprefix.Prefix, rName),
						tfjsonpath.New("role_definition"),
						fmt.Sprintf("%s_role_definition.%s", configprefix.Prefix, rName+"-1"),
						tfjsonpath.New("id"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_role_team_assignment.%s", configprefix.Prefix, rName),
						tfjsonpath.New("team"),
						fmt.Sprintf("%s_team.%s", configprefix.Prefix, rName+"-1"),
						tfjsonpath.New("id"),
						IdCompare,
					),
				},
			},
			{
				ResourceName:      fmt.Sprintf("%s_role_team_assignment.%s", configprefix.Prefix, rName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRoleTeamAssignmentResourceConfig(2, rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_role_team_assignment.%s", configprefix.Prefix, rName),
						tfjsonpath.New("object_id"),
						fmt.Sprintf("%s_organization.%s", configprefix.Prefix, rName+"-2"),
						tfjsonpath.New("id"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_role_team_assignment.%s", configprefix.Prefix, rName),
						tfjsonpath.New("role_definition"),
						fmt.Sprintf("%s_role_definition.%s", configprefix.Prefix, rName+"-2"),
						tfjsonpath.New("id"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_role_team_assignment.%s", configprefix.Prefix, rName),
						tfjsonpath.New("team"),
						fmt.Sprintf("%s_team.%s", configprefix.Prefix, rName+"-2"),
						tfjsonpath.New("id"),
						IdCompare,
					),
				},
			},
		},
	})
}

func testAccRoleTeamAssignmentResourceConfig(number int, rName string) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "%[4]s-%[3]d" {
  name        			= "%[2]s-%[3]d"
}
resource "%[1]s_role_definition" "%[4]s-%[3]d" {
  name         = "%[2]s"
  description  = "Test role definition"
  content_type = "shared.organization"
  permissions   = ["shared.view_organization"]
}
resource "%[1]s_team" "%[4]s-%[3]d" {
  name      = "%[2]s-%[3]d"
  organization   = %[1]s_organization.%[4]s-%[3]d.id
  description  = "%[2]s-%[3]d description"
}
resource "%[1]s_role_team_assignment" "%[4]s" {
  object_id       = %[1]s_organization.%[4]s-%[3]d.id
  role_definition = %[1]s_role_definition.%[4]s-%[3]d.id
  team            = %[1]s_team.%[4]s-%[3]d.id
}
`, configprefix.Prefix, acctest.RandString(5), number, rName)
}
