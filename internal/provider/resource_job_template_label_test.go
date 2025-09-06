package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/tfbrew/terraform-provider-aap/internal/configprefix"
)

func TestAccJobTemplateLabel_basic(t *testing.T) {
	testingJobTemplateName := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)

	stringListComparer := &compareStringInList{}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccJobTemplateLabel1Config(testingJobTemplateName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_label.test_label_1", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_job_template_label.test", configprefix.Prefix),
						tfjsonpath.New("label_ids"),
						stringListComparer,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_job_template.test", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_job_template_label.test", configprefix.Prefix),
						tfjsonpath.New("job_template_id"),
						compare.ValuesSame(),
					),
				},
			},
			{
				ResourceName:                         fmt.Sprintf("%s_job_template_label.test", configprefix.Prefix),
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    importStateJobTemplateID(fmt.Sprintf("%s_job_template_label.test", configprefix.Prefix)),
				ImportStateVerifyIdentifierAttribute: ("job_template_id"),
			},
		},
	})
}

func testAccJobTemplateLabel1Config(jobTemplateName string) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "test" {
  name        = "%[2]s"
}

resource "%[1]s_project" "test" {
	name         = "%[2]s"
	organization = %[1]s_organization.test.id
	allow_override = true
	scm_type = "git"
	scm_url = "https://github.com/fakerepo"
}	

resource "%[1]s_job_template" "test" {
	name = "%[3]s"
	playbook = "hello_world.yml"
	ask_inventory_on_launch = true
	project = %[1]s_project.test.id
}

resource "%[1]s_label" "test_label_1" {
	name = "testlabel1"
	organization = %[1]s_organization.test.id
}

resource "%[1]s_job_template_label" "test" {
	job_template_id = %[1]s_job_template.test.id
	label_ids = [%[1]s_label.test_label_1.id]
}

  `, configprefix.Prefix, acctest.RandString(5), jobTemplateName)
}
