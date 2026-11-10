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

func TestAccWkflwJobTemplJobNodeCredentialResource(t *testing.T) {
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
				Config: testAccWkflwJobTemplJobNodeCredentialResource1Config(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_workflow_job_template_job_node.%s", configprefix.Prefix, rName),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template_job_node_credential.%s", configprefix.Prefix, rName),
						tfjsonpath.New("id"),
						IdCompare,
					),
				},
			},
			{
				ResourceName:      fmt.Sprintf("%s_workflow_job_template_job_node_credential.%s", configprefix.Prefix, rName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccWkflwJobTemplJobNodeCredentialResource2Config(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_workflow_job_template_job_node.%s", configprefix.Prefix, rName),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template_job_node_credential.%s", configprefix.Prefix, rName),
						tfjsonpath.New("id"),
						IdCompare,
					),
				},
			},
		},
	})
}

func testAccWkflwJobTemplJobNodeCredentialResource1Config(rName string) string {
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
data "%[1]s_credential_type" "%[3]s" {
  name = "Machine"
  kind = "ssh"
}
resource "%[1]s_credential" "%[3]s" {
  name            = "%[2]s"
  description	    = "%[2]s"
  organization    = %[1]s_organization.%[3]s.id
  credential_type = data.%[1]s_credential_type.%[3]s.id
  inputs = jsonencode({
    "password" : "%[2]s",
    "username" : "%[2]s"
  })
}
resource "%[1]s_job_template" "%[3]s" {
  name        = "%[2]s"
  job_type    = "run"
  project     = %[1]s_project.%[3]s.id
  inventory   = %[1]s_inventory.%[3]s.id
  playbook    = "%[2]s"
  ask_credential_on_launch = true
  ask_inventory_on_launch  = true
}
resource "%[1]s_workflow_job_template" "%[3]s" {
  name                     = "%[2]s"
  inventory                = %[1]s_inventory.%[3]s.id
  organization             = %[1]s_organization.%[3]s.id
}
resource "%[1]s_workflow_job_template_job_node" "%[3]s" {
  unified_job_template     	= %[1]s_job_template.%[3]s.id
  workflow_job_template_id 	= %[1]s_workflow_job_template.%[3]s.id
  inventory 				        = %[1]s_inventory.%[3]s.id

}
resource "%[1]s_workflow_job_template_job_node_credential" "%[3]s" {
  credential_ids  = [ %[1]s_credential.%[3]s.id ]
  id              = %[1]s_workflow_job_template_job_node.%[3]s.id
}
  `, configprefix.Prefix, acctest.RandString(5), rName, rName+"a", rName+"b")
}

func testAccWkflwJobTemplJobNodeCredentialResource2Config(rName string) string {
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
data "%[1]s_credential_type" "%[4]s" {
  name = "Machine"
  kind = "ssh"
}
resource "%[1]s_credential" "%[4]s" {
  name            = "%[2]s"
  description	  = "%[2]s"
  organization    = %[1]s_organization.%[3]s.id
  credential_type = data.%[1]s_credential_type.%[4]s.id
  inputs = jsonencode({
    "password" : "%[2]s",
    "username" : "%[2]s"
  })
}
data "%[1]s_credential_type" "%[5]s" {
  name = "Amazon Web Services"
  kind = "cloud"
}
resource "%[1]s_credential" "%[5]s" {
  name            = "%[2]s"
  description	  = "%[2]s"
  organization    = %[1]s_organization.%[3]s.id
  credential_type = data.%[1]s_credential_type.%[5]s.id
  inputs = jsonencode({
    "password" : "%[2]s",
    "username" : "%[2]s"
  })
}
resource "%[1]s_job_template" "%[3]s" {
  name        = "%[2]s"
  job_type    = "run"
  project     = %[1]s_project.%[3]s.id
  playbook    = "%[2]s"
  inventory   = %[1]s_inventory.%[3]s.id
  ask_credential_on_launch = true
  ask_inventory_on_launch  = true
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
resource "%[1]s_workflow_job_template_job_node_credential" "%[3]s" {
  credential_ids  = [ %[1]s_credential.%[4]s.id, %[1]s_credential.%[5]s.id ]
  id = %[1]s_workflow_job_template_job_node.%[3]s.id
}
  `, configprefix.Prefix, acctest.RandString(5), rName, rName+"a", rName+"b")
}
