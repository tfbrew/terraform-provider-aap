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
	IdCompare := &compareTwoValuesAsStrings{}
	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccJobTemplateInstanceGroupResource1Config(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_job_template.test", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_job_template_instance_group.test", configprefix.Prefix),
						tfjsonpath.New("job_template_id"),
						IdCompare,
					),
				},
			},
			{
				ResourceName:                         fmt.Sprintf("%s_job_template_instance_group.test", configprefix.Prefix),
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    importStateJobTemplateID(fmt.Sprintf("%s_job_template_instance_group.test", configprefix.Prefix)),
				ImportStateVerifyIdentifierAttribute: ("job_template_id"),
			},
			{
				Config: testAccJobTemplateInstanceGroupResource2Config(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_job_template.test", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_job_template_instance_group.test", configprefix.Prefix),
						tfjsonpath.New("job_template_id"),
						IdCompare,
					),
				},
			},
		},
	})
}

func testAccJobTemplateInstanceGroupResource1Config() string {
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
resource "%[1]s_instance_group" "test" {
  name                       = "%[2]s"
}
resource "%[1]s_job_template" "test" {
  name        = "%[2]s"
  job_type    = "run"
  inventory   = %[1]s_inventory.test.id
  project     = %[1]s_project.test.id
  playbook    = "%[2]s"
}
resource "%[1]s_job_template_instance_group" "test" {
  instance_groups_ids  = [ %[1]s_instance_group.test.id ]
  job_template_id      = %[1]s_job_template.test.id
}
  `, configprefix.Prefix, acctest.RandString(5))
}

func testAccJobTemplateInstanceGroupResource2Config() string {
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
resource "%[1]s_instance_group" "test1" {
  name                       = "%[2]s-1"
}
resource "%[1]s_instance_group" "test2" {
  name                       = "%[2]s-2"
}
resource "%[1]s_job_template" "test" {
  name        = "%[2]s"
  job_type    = "run"
  inventory   = %[1]s_inventory.test.id
  project     = %[1]s_project.test.id
  playbook    = "%[2]s"
}
resource "%[1]s_job_template_instance_group" "test" {
  instance_groups_ids  = [ %[1]s_instance_group.test1.id, %[1]s_instance_group.test2.id ]
  job_template_id      = %[1]s_job_template.test.id
}
  `, configprefix.Prefix, acctest.RandString(5))
}
