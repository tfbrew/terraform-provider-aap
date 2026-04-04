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
	rName := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
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
				Config: testAccWkflwJobTemplJobNodeFailureResource1Config(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_workflow_job_template_job_node.%s", configprefix.Prefix, rName+"a"),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template_node_failure.%s", configprefix.Prefix, rName),
						tfjsonpath.New("id"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_workflow_job_template_job_node.%s", configprefix.Prefix, rName+"b"),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template_node_failure.%s", configprefix.Prefix, rName),
						tfjsonpath.New("failure_ids"),
						StringListCompare,
					),
				},
			},
			{
				ResourceName:      fmt.Sprintf("%s_workflow_job_template_node_failure.%s", configprefix.Prefix, rName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccWkflwJobTemplJobNodeFailureResource2Config(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_workflow_job_template_job_node.%s", configprefix.Prefix, rName+"c"),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template_node_failure.%s", configprefix.Prefix, rName),
						tfjsonpath.New("id"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_workflow_job_template_job_node.%s", configprefix.Prefix, rName+"d"),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template_node_failure.%s", configprefix.Prefix, rName),
						tfjsonpath.New("failure_ids"),
						StringListCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_workflow_job_template_job_node.%s", configprefix.Prefix, rName+"e"),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template_node_failure.%s", configprefix.Prefix, rName),
						tfjsonpath.New("failure_ids"),
						StringListCompare,
					),
				},
			},
		},
	})
}

func testAccWkflwJobTemplJobNodeFailureResource1Config(rName string) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "%[3]s" {
  name        = "%[2]s"
}
resource "%[1]s_inventory" "%[3]s" {
  name         = "%[2]s"
  organization = %[1]s_organization.%[3]s.id
}
resource "%[1]s_project" "%[3]s" {
  name         		= "%[2]s"
  organization 		= %[1]s_organization.%[3]s.id
  scm_type     		= "git"
  scm_url      		= "git@github.com:user/repo.git"
  allow_override 	= true
}
resource "%[1]s_job_template" "%[3]s" {
  name      				= "%[2]s"
  ask_inventory_on_launch 	= true
  project   				= %[1]s_project.%[3]s.id
  playbook  				= "test.yml"
}
resource "%[1]s_workflow_job_template" "%[3]s" {
  name                     = "%[2]s"
  inventory                = %[1]s_inventory.%[3]s.id
  organization             = %[1]s_organization.%[3]s.id
}
resource "%[1]s_workflow_job_template_job_node" "%[4]s" {
  unified_job_template     	= %[1]s_job_template.%[3]s.id
  workflow_job_template_id 	= %[1]s_workflow_job_template.%[3]s.id
  inventory 				= %[1]s_inventory.%[3]s.id
}
resource "%[1]s_workflow_job_template_job_node" "%[5]s" {
  unified_job_template     	= %[1]s_job_template.%[3]s.id
  workflow_job_template_id 	= %[1]s_workflow_job_template.%[3]s.id
  inventory 				= %[1]s_inventory.%[3]s.id
}
resource "%[1]s_workflow_job_template_node_failure" "%[3]s" {
  id              = %[1]s_workflow_job_template_job_node.%[4]s.id
  failure_ids = [%[1]s_workflow_job_template_job_node.%[5]s.id]
}
  `, configprefix.Prefix, acctest.RandString(5), rName, rName+"a", rName+"b")
}

func testAccWkflwJobTemplJobNodeFailureResource2Config(rName string) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "%[3]s" {
  name        = "%[2]s"
}
resource "%[1]s_inventory" "%[3]s" {
  name         = "%[2]s"
  organization = %[1]s_organization.%[3]s.id
}
resource "%[1]s_project" "%[3]s" {
  name         		= "%[2]s"
  organization 		= %[1]s_organization.%[3]s.id
  scm_type     		= "git"
  scm_url      		= "git@github.com:user/repo.git"
  allow_override 	= true
}
resource "%[1]s_job_template" "%[3]s" {
  name      = "%[2]s"
  inventory = %[1]s_inventory.%[3]s.id
  project   = %[1]s_project.%[3]s.id
  playbook  = "test.yml"
}
resource "%[1]s_workflow_job_template" "%[3]s" {
  name                     = "%[2]s"
  inventory                = %[1]s_inventory.%[3]s.id
  organization             = %[1]s_organization.%[3]s.id
}
resource "%[1]s_workflow_job_template_job_node" "%[4]s" {
  unified_job_template     	= %[1]s_job_template.%[3]s.id
  workflow_job_template_id 	= %[1]s_workflow_job_template.%[3]s.id
  inventory 				= %[1]s_inventory.%[3]s.id
}
resource "%[1]s_workflow_job_template_job_node" "%[5]s" {
  unified_job_template     	= %[1]s_job_template.%[3]s.id
  workflow_job_template_id 	= %[1]s_workflow_job_template.%[3]s.id
  inventory 				= %[1]s_inventory.%[3]s.id
}
resource "%[1]s_workflow_job_template_job_node" "%[6]s" {
  unified_job_template     	= %[1]s_job_template.%[3]s.id
  workflow_job_template_id 	= %[1]s_workflow_job_template.%[3]s.id
  inventory 				= %[1]s_inventory.%[3]s.id
}
resource "%[1]s_workflow_job_template_node_failure" "%[3]s" {
  id              = %[1]s_workflow_job_template_job_node.%[4]s.id
  failure_ids = [%[1]s_workflow_job_template_job_node.%[5]s.id, %[1]s_workflow_job_template_job_node.%[6]s.id]
}
  `, configprefix.Prefix, acctest.RandString(5), rName, rName+"c", rName+"d", rName+"e")
}
