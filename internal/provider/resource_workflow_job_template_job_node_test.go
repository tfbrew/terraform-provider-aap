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
				Config: testAccWkflwJobTemplJobNodeResource1Config(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_job_template.%s", configprefix.Prefix, rName),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template_job_node.%s", configprefix.Prefix, rName),
						tfjsonpath.New("unified_job_template"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_workflow_job_template.%s", configprefix.Prefix, rName),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template_job_node.%s", configprefix.Prefix, rName),
						tfjsonpath.New("workflow_job_template_id"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_inventory.%s", configprefix.Prefix, rName),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template_job_node.%s", configprefix.Prefix, rName),
						tfjsonpath.New("inventory"),
						IdCompare,
					),
				},
			},
			{
				ResourceName:      fmt.Sprintf("%s_workflow_job_template_job_node.%s", configprefix.Prefix, rName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccWkflwJobTemplJobNodeResource2Config(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_job_template.%s", configprefix.Prefix, rName),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template_job_node.%s", configprefix.Prefix, rName),
						tfjsonpath.New("unified_job_template"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_workflow_job_template.%s", configprefix.Prefix, rName),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template_job_node.%s", configprefix.Prefix, rName),
						tfjsonpath.New("workflow_job_template_id"),
						IdCompare,
					),
				},
			},
		},
	})
}

func testAccWkflwJobTemplJobNodeResource1Config(rName string) string {
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
resource "%[1]s_workflow_job_template_job_node" "%[3]s" {
  unified_job_template     	= %[1]s_job_template.%[3]s.id
  workflow_job_template_id 	= %[1]s_workflow_job_template.%[3]s.id
  inventory 				= %[1]s_inventory.%[3]s.id
}
  `, configprefix.Prefix, acctest.RandString(5), rName)
}

func testAccWkflwJobTemplJobNodeResource2Config(rName string) string {
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
resource "%[1]s_workflow_job_template_job_node" "%[3]s" {
  unified_job_template     = %[1]s_job_template.%[3]s.id
  workflow_job_template_id = %[1]s_workflow_job_template.%[3]s.id
}
  `, configprefix.Prefix, acctest.RandString(5), rName)
}
