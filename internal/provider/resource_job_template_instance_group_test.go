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

func TestAccJobTemplateInstanceGroupResource(t *testing.T) {
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
				Config: testAccJobTemplateInstanceGroupResource1Config(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_job_template.%s", configprefix.Prefix, rName),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_job_template_instance_group.%s", configprefix.Prefix, rName),
						tfjsonpath.New("job_template_id"),
						IdCompare,
					),
				},
			},
			{
				ResourceName:                         fmt.Sprintf("%s_job_template_instance_group.%s", configprefix.Prefix, rName),
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    importStateJobTemplateID(fmt.Sprintf("%s_job_template_instance_group.%s", configprefix.Prefix, rName)),
				ImportStateVerifyIdentifierAttribute: ("job_template_id"),
			},
			{
				Config: testAccJobTemplateInstanceGroupResource2Config(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_job_template.%s", configprefix.Prefix, rName),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_job_template_instance_group.%s", configprefix.Prefix, rName),
						tfjsonpath.New("job_template_id"),
						IdCompare,
					),
				},
			},
		},
	})
}

func testAccJobTemplateInstanceGroupResource1Config(rName string) string {
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
resource "%[1]s_instance_group" "%[3]s" {
  name                       = "%[2]s"
}
resource "%[1]s_job_template" "%[3]s" {
  name        = "%[2]s"
  job_type    = "run"
  inventory   = %[1]s_inventory.%[3]s.id
  project     = %[1]s_project.%[3]s.id
  playbook    = "%[2]s"
}
resource "%[1]s_job_template_instance_group" "%[3]s" {
  instance_groups_ids  = [ %[1]s_instance_group.%[3]s.id ]
  job_template_id      = %[1]s_job_template.%[3]s.id
}
  `, configprefix.Prefix, acctest.RandString(5), rName)
}

func testAccJobTemplateInstanceGroupResource2Config(rName string) string {
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
resource "%[1]s_instance_group" "%[4]s" {
  name                       = "%[2]s-1"
}
resource "%[1]s_instance_group" "%[5]s" {
  name                       = "%[2]s-2"
}
resource "%[1]s_job_template" "%[3]s" {
  name        = "%[2]s"
  job_type    = "run"
  inventory   = %[1]s_inventory.%[3]s.id
  project     = %[1]s_project.%[3]s.id
  playbook    = "%[2]s"
}
resource "%[1]s_job_template_instance_group" "%[3]s" {
  instance_groups_ids  = [ %[1]s_instance_group.%[4]s.id, %[1]s_instance_group.%[5]s.id ]
  job_template_id      = %[1]s_job_template.%[3]s.id
}
  `, configprefix.Prefix, acctest.RandString(5), rName, rName+"a", rName+"b")
}
