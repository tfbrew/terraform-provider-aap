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

func TestAccJobTemplateResource(t *testing.T) {
	rName := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	IdCompare := &compareTwoValuesAsStrings{}
	resource1 := JobTemplateAPIModel{
		Name:        "test-job-template-" + acctest.RandString(5),
		Description: "test description 1",
		JobType:     "run",
		Playbook:    "test.yml",
	}
	resource2 := JobTemplateAPIModel{
		Name:        "test-job-template-" + acctest.RandString(5),
		Description: "test description 2",
		JobType:     "check",
		Playbook:    "test2.yml",
	}
	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccJobTemplateResourceConfig(resource1, rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_job_template.%s", configprefix.Prefix, rName),
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource1.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_job_template.%s", configprefix.Prefix, rName),
						tfjsonpath.New("description"),
						knownvalue.StringExact(resource1.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_job_template.%s", configprefix.Prefix, rName),
						tfjsonpath.New("job_type"),
						knownvalue.StringExact(resource1.JobType),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_job_template.%s", configprefix.Prefix, rName),
						tfjsonpath.New("playbook"),
						knownvalue.StringExact(resource1.Playbook),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_inventory.%s", configprefix.Prefix, rName),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_job_template.%s", configprefix.Prefix, rName),
						tfjsonpath.New("inventory"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_project.%s", configprefix.Prefix, rName),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_job_template.%s", configprefix.Prefix, rName),
						tfjsonpath.New("project"),
						IdCompare,
					),
				},
			},
			{
				ResourceName:      fmt.Sprintf("%s_job_template.%s", configprefix.Prefix, rName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccJobTemplateResourceConfig(resource2, rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_job_template.%s", configprefix.Prefix, rName),
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource2.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_job_template.%s", configprefix.Prefix, rName),
						tfjsonpath.New("description"),
						knownvalue.StringExact(resource2.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_job_template.%s", configprefix.Prefix, rName),
						tfjsonpath.New("job_type"),
						knownvalue.StringExact(resource2.JobType),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_job_template.%s", configprefix.Prefix, rName),
						tfjsonpath.New("playbook"),
						knownvalue.StringExact(resource2.Playbook),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_job_template.test", configprefix.Prefix),
						tfjsonpath.New("extra_vars"),
						knownvalue.StringExact(""),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_inventory.%s", configprefix.Prefix, rName),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_job_template.%s", configprefix.Prefix, rName),
						tfjsonpath.New("inventory"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_project.%s", configprefix.Prefix, rName),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_job_template.%s", configprefix.Prefix, rName),
						tfjsonpath.New("project"),
						IdCompare,
					),
				},
			},
		},
	})
}

func testAccJobTemplateResourceConfig(resource JobTemplateAPIModel, rName string) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "%[7]s" {
  name        = "%[2]s"
}

resource "%[1]s_inventory" "%[7]s" {
  name         = "%[2]s"
  organization = %[1]s_organization.%[7]s.id
}

resource "%[1]s_project" "%[7]s" {
  name         		= "%[2]s"
  organization 		= %[1]s_organization.%[7]s.id
  scm_type     		= "git"
  scm_url      		= "git@github.com:user/repo.git"
  allow_override 	= true
}

resource "%[1]s_job_template" "%[7]s" {
  name        = "%[3]s"
  description = "%[4]s"
  job_type    = "%[5]s"
  inventory   = %[1]s_inventory.%[7]s.id
  project     = %[1]s_project.%[7]s.id
  playbook    = "%[6]s"
}
  `, configprefix.Prefix, acctest.RandString(5), resource.Name, resource.Description, resource.JobType, resource.Playbook, rName)
}
