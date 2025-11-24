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

func TestAccWorkflowJobTemplateResource(t *testing.T) {
	IdCompare := &compareTwoValuesAsStrings{}
	resource1 := WorkflowJobTemplateAPIModel{
		Name: "test-workflow-job-template" + acctest.RandString(5),
	}
	resource2 := WorkflowJobTemplateAPIModel{
		Name:        "test-workflow-job-template" + acctest.RandString(5),
		Description: "test description2",
	}
	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowJobTemplateResource1Config(resource1),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_workflow_job_template.test", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource1.Name),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_inventory.test", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template.test", configprefix.Prefix),
						tfjsonpath.New("inventory"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_organization.test", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template.test", configprefix.Prefix),
						tfjsonpath.New("organization"),
						IdCompare,
					),
				},
			},
			{
				ResourceName:      fmt.Sprintf("%s_workflow_job_template.test", configprefix.Prefix),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccWorkflowJobTemplateResource2Config(resource2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_workflow_job_template.test", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource2.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_workflow_job_template.test", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(resource2.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_workflow_job_template.test", configprefix.Prefix),
						tfjsonpath.New("extra_vars"),
						knownvalue.Null(),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_inventory.test", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template.test", configprefix.Prefix),
						tfjsonpath.New("inventory"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_organization.test", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template.test", configprefix.Prefix),
						tfjsonpath.New("organization"),
						IdCompare,
					),
				},
			},
		},
	})
}

func testAccWorkflowJobTemplateResource1Config(resource WorkflowJobTemplateAPIModel) string {
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
  name         = "%[3]s"
  inventory    = %[1]s_inventory.test.id
  organization = %[1]s_organization.test.id
}
  `, configprefix.Prefix, acctest.RandString(5), resource.Name)
}

func testAccWorkflowJobTemplateResource2Config(resource WorkflowJobTemplateAPIModel) string {
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
  name         = "%[3]s"
  description  = "%[4]s"
  inventory    = %[1]s_inventory.test.id
  organization = %[1]s_organization.test.id
}
  `, configprefix.Prefix, acctest.RandString(5), resource.Name, resource.Description)
}
