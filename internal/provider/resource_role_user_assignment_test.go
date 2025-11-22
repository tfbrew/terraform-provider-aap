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

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleUserAssignmentResourceConfig(1),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_role_user_assignment.test", configprefix.Prefix),
						tfjsonpath.New("object_id"),
						fmt.Sprintf("%s_project.test-1", configprefix.Prefix),
						tfjsonpath.New("id"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_role_user_assignment.test", configprefix.Prefix),
						tfjsonpath.New("role_definition"),
						fmt.Sprintf("%s_role_definition.test-1", configprefix.Prefix),
						tfjsonpath.New("id"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_role_user_assignment.test", configprefix.Prefix),
						tfjsonpath.New("user"),
						fmt.Sprintf("%s_user.test-1", configprefix.Prefix),
						tfjsonpath.New("id"),
						IdCompare,
					),
				},
			},
			{
				ResourceName:      fmt.Sprintf("%s_role_user_assignment.test", configprefix.Prefix),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRoleUserAssignmentResourceConfig(2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_role_user_assignment.test", configprefix.Prefix),
						tfjsonpath.New("object_id"),
						fmt.Sprintf("%s_project.test-2", configprefix.Prefix),
						tfjsonpath.New("id"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_role_user_assignment.test", configprefix.Prefix),
						tfjsonpath.New("role_definition"),
						fmt.Sprintf("%s_role_definition.test-2", configprefix.Prefix),
						tfjsonpath.New("id"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_role_user_assignment.test", configprefix.Prefix),
						tfjsonpath.New("user"),
						fmt.Sprintf("%s_user.test-2", configprefix.Prefix),
						tfjsonpath.New("id"),
						IdCompare,
					),
				},
			},
		},
	})
}

func testAccRoleUserAssignmentResourceConfig(number int) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "test-%[3]d" {
  name        			= "%[2]s-%[3]d"
}
resource "%[1]s_project" "test-%[3]d" {
  name        			= "%[2]s-%[3]d"
  scm_type			    = "git"
  scm_url			    = "https://github.com/"
  organization  		= %[1]s_organization.test-%[3]d.id
}
resource "%[1]s_role_definition" "test-%[3]d" {
  name         = "%[2]s"
  description  = "Test role definition"
  content_type = "awx.project"
  permissions   = ["awx.use_project", "awx.view_project"]
}
resource "%[1]s_user" "test-%[3]d" {
  username      = "%[2]s-%[3]d"
  first_name 	= "%[2]s-%[3]d"
  last_name 	= "%[2]s-%[3]d"
  email			= "%[2]s-%[3]d@example.com"
  password 		= "%[2]s-%[3]d-password"
}
resource "%[1]s_role_user_assignment" "test" {
  object_id       = %[1]s_project.test-%[3]d.id
  role_definition = %[1]s_role_definition.test-%[3]d.id
  user            = %[1]s_user.test-%[3]d.id
}
`, configprefix.Prefix, acctest.RandString(5), number)
}
