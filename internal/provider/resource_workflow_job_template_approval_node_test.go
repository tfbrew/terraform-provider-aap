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

func TestAccWkflwJobTemplApprovalNodeResource(t *testing.T) {
	IdCompare := &compareTwoValuesAsStrings{}
	nodeName := acctest.RandString(5)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWkflwJobTemplApprvlNodeResource1Config(nodeName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_workflow_job_template.test", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template_approval_node.test", configprefix.Prefix),
						tfjsonpath.New("workflow_job_template_id"),
						IdCompare,
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_workflow_job_template_approval_node.test", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(nodeName),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_workflow_job_template_approval_node.test", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact("a description for testing"),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_workflow_job_template_approval_node.test", configprefix.Prefix),
						tfjsonpath.New("timeout"),
						knownvalue.Int32Exact(360),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_workflow_job_template_approval_node.test_timeoutdefault", configprefix.Prefix),
						tfjsonpath.New("timeout"),
						knownvalue.Int32Exact(0),
					),
				},
			},
			{
				ResourceName:      fmt.Sprintf("%s_workflow_job_template_approval_node.test", configprefix.Prefix),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      fmt.Sprintf("%s_workflow_job_template_approval_node.test_timeoutdefault", configprefix.Prefix),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccWkflwJobTemplApprvlNodeResource1Config(nodeName string) string {
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

resource "%[1]s_workflow_job_template_approval_node" "test" {
  workflow_job_template_id 	= %[1]s_workflow_job_template.test.id
  name = "%[3]s"
  description = "a description for testing"
  timeout = 360
}

resource "%[1]s_workflow_job_template_approval_node" "test_timeoutdefault" {
  workflow_job_template_id 	= %[1]s_workflow_job_template.test.id
  name = "%[4]s"
  description = "a description for testing"
}
  `, configprefix.Prefix, acctest.RandString(5), nodeName, nodeName+"1")
}
