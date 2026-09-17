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

func TestAccWorkflowJobTemplateLabel(t *testing.T) {
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
				Config: testAccWorkflowJobTemplateLabelConfig(testingJobTemplateName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_label.%s", configprefix.Prefix, testingJobTemplateName),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template_label.%s", configprefix.Prefix, testingJobTemplateName),
						tfjsonpath.New("label_ids"),
						stringListComparer,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_workflow_job_template.%s", configprefix.Prefix, testingJobTemplateName),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template_label.%s", configprefix.Prefix, testingJobTemplateName),
						tfjsonpath.New("workflow_job_template_id"),
						compare.ValuesSame(),
					),
				},
			},
			{
				ResourceName:                         fmt.Sprintf("%s_workflow_job_template_label.%s", configprefix.Prefix, testingJobTemplateName),
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    importStateWorkflowJobTemplateID(fmt.Sprintf("%s_workflow_job_template_label.%s", configprefix.Prefix, testingJobTemplateName)),
				ImportStateVerifyIdentifierAttribute: ("workflow_job_template_id"),
			},
		},
	})
}

func testAccWorkflowJobTemplateLabelConfig(jobTemplateName string) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "%[3]s" {
  name        = "%[2]s"
}	

resource "%[1]s_workflow_job_template" "%[3]s" {
  name         = "%[3]s"
  organization = %[1]s_organization.%[3]s.id
}

resource "%[1]s_label" "%[3]s" {
	name = "testlabel1"
	organization = %[1]s_organization.%[3]s.id
}

resource "%[1]s_workflow_job_template_label" "%[3]s" {
	workflow_job_template_id = %[1]s_workflow_job_template.%[3]s.id
	label_ids = [%[1]s_label.%[3]s.id]
}

  `, configprefix.Prefix, acctest.RandString(5), jobTemplateName)
}
