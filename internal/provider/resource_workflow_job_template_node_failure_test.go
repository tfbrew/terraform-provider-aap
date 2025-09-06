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

func TestAccWkflwJobTemplJobNodeFailureResource(t *testing.T) {
	IdCompare := &compareTwoValuesAsStrings{}
	StringListCompare := &compareStringInList{}
	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWkflwJobTemplJobNodeFailureResource1Config(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_workflow_job_template_job_node.test1", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template_node_failure.test", configprefix.Prefix),
						tfjsonpath.New("id"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_workflow_job_template_job_node.test2", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template_node_failure.test", configprefix.Prefix),
						tfjsonpath.New("failure_ids"),
						StringListCompare,
					),
				},
			},
			{
				ResourceName:      fmt.Sprintf("%s_workflow_job_template_node_failure.test", configprefix.Prefix),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccWkflwJobTemplJobNodeFailureResource2Config(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_workflow_job_template_job_node.test3", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template_node_failure.test", configprefix.Prefix),
						tfjsonpath.New("id"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_workflow_job_template_job_node.test4", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template_node_failure.test", configprefix.Prefix),
						tfjsonpath.New("failure_ids"),
						StringListCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_workflow_job_template_job_node.test5", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template_node_failure.test", configprefix.Prefix),
						tfjsonpath.New("failure_ids"),
						StringListCompare,
					),
				},
			},
		},
	})
}

func testAccWkflwJobTemplJobNodeFailureResource1Config() string {
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
resource "%[1]s_workflow_job_template_job_node" "test1" {
  unified_job_template     	= %[1]s_job_template.test.id
  workflow_job_template_id 	= %[1]s_workflow_job_template.test.id
  inventory 				= %[1]s_inventory.test.id
}
resource "%[1]s_workflow_job_template_job_node" "test2" {
  unified_job_template     	= %[1]s_job_template.test.id
  workflow_job_template_id 	= %[1]s_workflow_job_template.test.id
  inventory 				= %[1]s_inventory.test.id
}
resource "%[1]s_workflow_job_template_node_failure" "test" {
  id              = %[1]s_workflow_job_template_job_node.test1.id
  failure_ids = [%[1]s_workflow_job_template_job_node.test2.id]
}
  `, configprefix.Prefix, acctest.RandString(5))
}

func testAccWkflwJobTemplJobNodeFailureResource2Config() string {
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
resource "%[1]s_workflow_job_template_job_node" "test3" {
  unified_job_template     	= %[1]s_job_template.test.id
  workflow_job_template_id 	= %[1]s_workflow_job_template.test.id
  inventory 				= %[1]s_inventory.test.id
}
resource "%[1]s_workflow_job_template_job_node" "test4" {
  unified_job_template     	= %[1]s_job_template.test.id
  workflow_job_template_id 	= %[1]s_workflow_job_template.test.id
  inventory 				= %[1]s_inventory.test.id
}
resource "%[1]s_workflow_job_template_job_node" "test5" {
  unified_job_template     	= %[1]s_job_template.test.id
  workflow_job_template_id 	= %[1]s_workflow_job_template.test.id
  inventory 				= %[1]s_inventory.test.id
}
resource "%[1]s_workflow_job_template_node_failure" "test" {
  id              = %[1]s_workflow_job_template_job_node.test3.id
  failure_ids = [%[1]s_workflow_job_template_job_node.test4.id, %[1]s_workflow_job_template_job_node.test5.id]
}
  `, configprefix.Prefix, acctest.RandString(5))
}
