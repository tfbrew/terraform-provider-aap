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
	rName := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWkflwJobTemplApprvlNodeResource1Config(nodeName, rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_workflow_job_template.%s", configprefix.Prefix, rName),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template_approval_node.%s", configprefix.Prefix, rName),
						tfjsonpath.New("workflow_job_template_id"),
						IdCompare,
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_workflow_job_template_approval_node.%s", configprefix.Prefix, rName),
						tfjsonpath.New("name"),
						knownvalue.StringExact(nodeName),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_workflow_job_template_approval_node.%s", configprefix.Prefix, rName),
						tfjsonpath.New("description"),
						knownvalue.StringExact("a description for testing"),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_workflow_job_template_approval_node.%s", configprefix.Prefix, rName),
						tfjsonpath.New("timeout"),
						knownvalue.Int32Exact(360),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_workflow_job_template_approval_node.%s", configprefix.Prefix, rName+"d"),
						tfjsonpath.New("timeout"),
						knownvalue.Int32Exact(0),
					),
				},
			},
			{
				ResourceName:      fmt.Sprintf("%s_workflow_job_template_approval_node.%s", configprefix.Prefix, rName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      fmt.Sprintf("%s_workflow_job_template_approval_node.%s", configprefix.Prefix, rName+"d"),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccWkflwJobTemplApprvlNodeResource1Config(nodeName string, rName string) string {
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
  name      				= "%[2]s"
  ask_inventory_on_launch 	= true
  project   				= %[1]s_project.%[5]s.id
  playbook  				= "test.yml"
}
resource "%[1]s_workflow_job_template" "%[5]s" {
  name                     = "%[2]s"
  inventory                = %[1]s_inventory.%[5]s.id
  organization             = %[1]s_organization.%[5]s.id
}

resource "%[1]s_workflow_job_template_approval_node" "%[5]s" {
  workflow_job_template_id 	= %[1]s_workflow_job_template.%[5]s.id
  name = "%[3]s"
  description = "a description for testing"
  timeout = 360
}

resource "%[1]s_workflow_job_template_approval_node" "%[6]s" {
  workflow_job_template_id 	= %[1]s_workflow_job_template.%[5]s.id
  name = "%[4]s"
  description = "a description for testing"
}
  `, configprefix.Prefix, acctest.RandString(5), nodeName, nodeName+"1", rName, rName+"d")
}
