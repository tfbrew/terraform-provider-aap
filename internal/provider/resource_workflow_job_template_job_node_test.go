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

func TestAccWkflwJobTemplJobNodeResource(t *testing.T) {
	IdCompare := &compareTwoValuesAsStrings{}
	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWkflwJobTemplJobNodeResource1Config(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_job_template.test", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template_job_node.test", configprefix.Prefix),
						tfjsonpath.New("unified_job_template"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_workflow_job_template.test", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template_job_node.test", configprefix.Prefix),
						tfjsonpath.New("workflow_job_template_id"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_inventory.test", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template_job_node.test", configprefix.Prefix),
						tfjsonpath.New("inventory"),
						IdCompare,
					),
				},
			},
			{
				ResourceName:      fmt.Sprintf("%s_workflow_job_template_job_node.test", configprefix.Prefix),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccWkflwJobTemplJobNodeResource2Config(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_job_template.test", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template_job_node.test", configprefix.Prefix),
						tfjsonpath.New("unified_job_template"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_workflow_job_template.test", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template_job_node.test", configprefix.Prefix),
						tfjsonpath.New("workflow_job_template_id"),
						IdCompare,
					),
				},
			},
		},
	})
}

func testAccWkflwJobTemplJobNodeResource1Config() string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "test" {
  name        = "%[2]s"
}
resource "%[1]s_inventory" "test" {
  name         = "%[2]s"
  organization = %[1]s_organization.test.id
}
resource "%[1]s_project" "test" {
  name         		= "%[2]s"
  organization 		= %[1]s_organization.test.id
  scm_type     		= "git"
  scm_url      		= "git@github.com:user/repo.git"
  allow_override 	= true
}
resource "%[1]s_job_template" "test" {
  name      				= "%[2]s"
  ask_inventory_on_launch 	= true
  project   				= %[1]s_project.test.id
  playbook  				= "test.yml"
}
resource "%[1]s_workflow_job_template" "test" {
  name                     = "%[2]s"
  inventory                = %[1]s_inventory.test.id
  organization             = %[1]s_organization.test.id
}
resource "%[1]s_workflow_job_template_job_node" "test" {
  unified_job_template     	= %[1]s_job_template.test.id
  workflow_job_template_id 	= %[1]s_workflow_job_template.test.id
  inventory 				= %[1]s_inventory.test.id
}
  `, configprefix.Prefix, acctest.RandString(5))
}

func testAccWkflwJobTemplJobNodeResource2Config() string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "test" {
  name        = "%[2]s"
}
resource "%[1]s_inventory" "test" {
  name         = "%[2]s"
  organization = %[1]s_organization.test.id
}
resource "%[1]s_project" "test" {
  name         		= "%[2]s"
  organization 		= %[1]s_organization.test.id
  scm_type     		= "git"
  scm_url      		= "git@github.com:user/repo.git"
  allow_override 	= true
}
resource "%[1]s_job_template" "test" {
  name      = "%[2]s"
  inventory = %[1]s_inventory.test.id
  project   = %[1]s_project.test.id
  playbook  = "test.yml"
}
resource "%[1]s_workflow_job_template" "test" {
  name                     = "%[2]s"
  inventory                = %[1]s_inventory.test.id
  organization             = %[1]s_organization.test.id
}
resource "%[1]s_workflow_job_template_job_node" "test" {
  unified_job_template     = %[1]s_job_template.test.id
  workflow_job_template_id = %[1]s_workflow_job_template.test.id
}
  `, configprefix.Prefix, acctest.RandString(5))
}
