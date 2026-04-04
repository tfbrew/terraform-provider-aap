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
	rName := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	rName2 := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
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
				Config: testAccWorkflowJobTemplateResource1Config(resource1, rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_workflow_job_template.%s", configprefix.Prefix, rName),
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource1.Name),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_inventory.%s", configprefix.Prefix, rName),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template.%s", configprefix.Prefix, rName),
						tfjsonpath.New("inventory"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_organization.%s", configprefix.Prefix, rName),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template.%s", configprefix.Prefix, rName),
						tfjsonpath.New("organization"),
						IdCompare,
					),
				},
			},
			{
				ResourceName:      fmt.Sprintf("%s_workflow_job_template.%s", configprefix.Prefix, rName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccWorkflowJobTemplateResource2Config(resource2, rName2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_workflow_job_template.%s", configprefix.Prefix, rName2),
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource2.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_workflow_job_template.%s", configprefix.Prefix, rName2),
						tfjsonpath.New("description"),
						knownvalue.StringExact(resource2.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_workflow_job_template.test", configprefix.Prefix),
						tfjsonpath.New("extra_vars"),
						knownvalue.Null(),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_inventory.%s", configprefix.Prefix, rName2),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template.%s", configprefix.Prefix, rName2),
						tfjsonpath.New("inventory"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_organization.%s", configprefix.Prefix, rName2),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template.%s", configprefix.Prefix, rName2),
						tfjsonpath.New("organization"),
						IdCompare,
					),
				},
			},
		},
	})
}

func testAccWorkflowJobTemplateResource1Config(resource WorkflowJobTemplateAPIModel, rName string) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "%[4]s" {
  name        = "%[2]s"
}
resource "%[1]s_inventory" "%[4]s" {
  name         = "%[2]s"
  organization = %[1]s_organization.%[4]s.id
}
resource "%[1]s_project" "%[4]s" {
  name         		= "%[2]s"
  organization 		= %[1]s_organization.%[4]s.id
  scm_type     		= "git"
  scm_url      		= "git@github.com:user/repo.git"
  allow_override 	= true
}
resource "%[1]s_job_template" "%[4]s" {
  name      				= "%[2]s"
  ask_inventory_on_launch 	= true
  project   				= %[1]s_project.%[4]s.id
  playbook  				= "test.yml"
}
resource "%[1]s_workflow_job_template" "%[4]s" {
  name         = "%[3]s"
  inventory    = %[1]s_inventory.%[4]s.id
  organization = %[1]s_organization.%[4]s.id
}
  `, configprefix.Prefix, acctest.RandString(5), resource.Name, rName)
}

func testAccWorkflowJobTemplateResource2Config(resource WorkflowJobTemplateAPIModel, rName string) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "%[5]s" {
  name        = "%[2]s"
}
resource "%[1]s_inventory" "%[5]s" {
  name         = "%[2]s"
  organization = %[1]s_organization.%[5]s.id
}
resource "%[1]s_project" "%[5]s" {
  name         		= "%[2]s"
  organization 		= %[1]s_organization.%[5]s.id
  scm_type     		= "git"
  scm_url      		= "git@github.com:user/repo.git"
  allow_override 	= true
}
resource "%[1]s_job_template" "%[5]s" {
  name      = "%[2]s"
  inventory = %[1]s_inventory.%[5]s.id
  project   = %[1]s_project.%[5]s.id
  playbook  = "test.yml"
}
resource "%[1]s_workflow_job_template" "%[5]s" {
  name         = "%[3]s"
  description  = "%[4]s"
  inventory    = %[1]s_inventory.%[5]s.id
  organization = %[1]s_organization.%[5]s.id
}
  `, configprefix.Prefix, acctest.RandString(5), resource.Name, resource.Description, rName)
}
