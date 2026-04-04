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

func TestAccRoleUserAssignmentResource(t *testing.T) {
	IdCompare := &compareTwoValuesAsStrings{}
	rName := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleUserAssignmentResourceConfig(1, rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_role_user_assignment.%s-1", configprefix.Prefix, rName),
						tfjsonpath.New("object_id"),
						fmt.Sprintf("%s_organization.%s-1", configprefix.Prefix, rName),
						tfjsonpath.New("id"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_role_user_assignment.%s-1", configprefix.Prefix, rName),
						tfjsonpath.New("role_definition"),
						fmt.Sprintf("%s_role_definition.%s-1", configprefix.Prefix, rName),
						tfjsonpath.New("id"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_role_user_assignment.%s-1", configprefix.Prefix, rName),
						tfjsonpath.New("user"),
						fmt.Sprintf("%s_user.%s-1", configprefix.Prefix, rName),
						tfjsonpath.New("id"),
						IdCompare,
					),
				},
			},
			{
				ResourceName:      fmt.Sprintf("%s_role_user_assignment.%s-1", configprefix.Prefix, rName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRoleUserAssignmentResourceConfig(2, rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_role_user_assignment.%s-2", configprefix.Prefix, rName),
						tfjsonpath.New("object_id"),
						fmt.Sprintf("%s_organization.%s-2", configprefix.Prefix, rName),
						tfjsonpath.New("id"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_role_user_assignment.%s-2", configprefix.Prefix, rName),
						tfjsonpath.New("role_definition"),
						fmt.Sprintf("%s_role_definition.%s-2", configprefix.Prefix, rName),
						tfjsonpath.New("id"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_role_user_assignment.%s-2", configprefix.Prefix, rName),
						tfjsonpath.New("user"),
						fmt.Sprintf("%s_user.%s-2", configprefix.Prefix, rName),
						tfjsonpath.New("id"),
						IdCompare,
					),
				},
			},
		},
	})
}

func testAccRoleUserAssignmentResourceConfig(number int, rName string) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "%[4]s-%[3]d" {
  name        			= "%[2]s-%[3]d"
}
resource "%[1]s_role_definition" "%[4]s-%[3]d" {
  name         = "%[2]s-%[3]d"
  description  = "Test role definition"
  content_type = "shared.organization"
  permissions   = ["shared.member_organization", "shared.view_organization"]
}
resource "%[1]s_user" "%[4]s-%[3]d" {
  username      = "%[2]s-%[3]d"
  first_name 	= "%[2]s-%[3]d"
  last_name 	= "%[2]s-%[3]d"
  email			= "%[2]s-%[3]d@example.com"
  password 		= "%[2]s-%[3]d-password"
}
resource "%[1]s_role_user_assignment" "%[4]s-%[3]d" {
  object_id       = %[1]s_organization.%[4]s-%[3]d.id
  role_definition = %[1]s_role_definition.%[4]s-%[3]d.id
  user            = %[1]s_user.%[4]s-%[3]d.id
}
`, configprefix.Prefix, acctest.RandString(5), number, rName)
}
