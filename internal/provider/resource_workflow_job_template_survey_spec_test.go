package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/tfbrew/terraform-provider-aap/internal/configprefix"
)

func TestAccWkflwJobTemplateSurveySpec(t *testing.T) {
	rName := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	wkflkJtName := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWkflwJobTemplateSurveySpecConfig(wkflkJtName, rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_workflow_job_template_survey_spec.%s", configprefix.Prefix, rName),
						tfjsonpath.New("description"),
						knownvalue.StringExact("test description"),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_workflow_job_template.%s", configprefix.Prefix, rName),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template_survey_spec.%s", configprefix.Prefix, rName),
						tfjsonpath.New("id"),
						compare.ValuesSame(),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_workflow_job_template_survey_spec.%s", configprefix.Prefix, rName),
						tfjsonpath.New("name"),
						knownvalue.StringExact(""),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_workflow_job_template_survey_spec.%s", configprefix.Prefix, rName),
						tfjsonpath.New("spec"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"max":                  knownvalue.Int32Exact(1024),
								"min":                  knownvalue.Int32Exact(0),
								"type":                 knownvalue.StringExact("multiselect"),
								"question_name":        knownvalue.StringExact("test_question_1"),
								"question_description": knownvalue.StringExact("test question 1"),
								"variable":             knownvalue.StringExact("testvar1"),
								"required":             knownvalue.Bool(true),
								"default":              knownvalue.StringExact("choice2\nchoice3"),
								"choices": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.StringExact("choice1"),
									knownvalue.StringExact("choice2"),
									knownvalue.StringExact("choice3"),
								}),
							}),
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"max":                  knownvalue.Int32Exact(1024),
								"min":                  knownvalue.Int32Exact(0),
								"type":                 knownvalue.StringExact("multiplechoice"),
								"question_name":        knownvalue.StringExact("test_question_2"),
								"question_description": knownvalue.StringExact("test question 2"),
								"variable":             knownvalue.StringExact("testvar2"),
								"required":             knownvalue.Bool(true),
								"default":              knownvalue.StringExact("status"),
								"choices": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.StringExact("stop"),
									knownvalue.StringExact("start"),
									knownvalue.StringExact("status"),
									knownvalue.StringExact("restart"),
								}),
							}),
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"max":                  knownvalue.Int32Exact(999),
								"min":                  knownvalue.Int32Exact(1),
								"type":                 knownvalue.StringExact("integer"),
								"question_name":        knownvalue.StringExact("test question 3"),
								"question_description": knownvalue.StringExact("test question 3"),
								"variable":             knownvalue.StringExact("test_3_var"),
								"required":             knownvalue.Bool(true),
								"default":              knownvalue.StringExact("15"),
								"choices":              knownvalue.Null(),
							}),
						}),
					),
				},
			},
			{
				ResourceName:      fmt.Sprintf("%s_workflow_job_template_survey_spec.%s", configprefix.Prefix, rName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccWkflwJobTemplateSurveySpecConfig(workflow_template_name string, rName string) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "%[5]s" {
	name = "%[2]s"
}

resource "%[1]s_inventory" "%[5]s" {
  name         = "%[2]s"
  organization = %[1]s_organization.%[5]s.id
}

resource "%[1]s_project" "%[5]s" {
	name = "%[3]s"
	organization = %[1]s_organization.%[5]s.id
	allow_override = true
	scm_type = "git"
	scm_url = "fake"
}

resource "%[1]s_job_template" "%[5]s" {
  job_type  = "run"
  name      = "%[4]s"
  ask_inventory_on_launch = true
  project   = %[1]s_project.%[5]s.id
  playbook  = "hello_world.yml"
}

resource "%[1]s_workflow_job_template" "%[5]s" {
  name                     = "%[2]s"
  inventory                = %[1]s_inventory.%[5]s.id
  organization             = %[1]s_organization.%[5]s.id
}

resource "%[1]s_workflow_job_template_survey_spec" "%[5]s" {
  description = "test description"
  id          = %[1]s_workflow_job_template.%[5]s.id
  name        = ""
  spec = [
    {
      choices              = ["choice1", "choice2", "choice3"]
      default              = "choice2\nchoice3"
      max                  = 1024
      min                  = 0
      question_description = "test question 1"
      question_name        = "test_question_1"
      required             = true
      type                 = "multiselect"
      variable             = "testvar1"
    },
    {
      choices              = ["stop", "start", "status", "restart"]
      default              = "status"
      max                  = 1024
      min                  = 0
      question_description = "test question 2"
      question_name        = "test_question_2"
      required             = true
      type                 = "multiplechoice"
      variable             = "testvar2"
    },
    {
      default              = jsonencode(15)
      max                  = 999
      min                  = 1
      question_description = "test question 3"
      question_name        = "test question 3"
      required             = true
      type                 = "integer"
      variable             = "test_3_var"
    },
  ]
}
`, configprefix.Prefix, acctest.RandStringFromCharSet(5, acctest.CharSetAlpha), acctest.RandStringFromCharSet(5, acctest.CharSetAlpha), workflow_template_name, rName)
}
