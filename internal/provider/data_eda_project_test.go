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

func TestAccEdaProjectDataSource(t *testing.T) {
	IdCompare := &compareTwoValuesAsStrings{}
	edaproject1 := EdaProjectAPIModel{
		Name:        "test-project-" + acctest.RandString(5),
		Description: "Initial test git project",
		ScmType:     "git",
		Url:         "https://github.com/example/repo.git",
	}

	edaproject2 := EdaProjectAPIModel{
		Name:        "test-project-" + acctest.RandString(5),
		Description: "Updated test git project",
		ScmType:     "git",
		Url:         "https://github.com/example/updated-repo.git",
		VerifySsl:   false,
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEdaProjectDataSourceConfig(edaproject1),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_eda_project.test-id", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(edaproject1.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_eda_project.test-id", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(edaproject1.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_eda_project.test-id", configprefix.Prefix),
						tfjsonpath.New("scm_type"),
						knownvalue.StringExact(edaproject1.ScmType),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_eda_project.test-id", configprefix.Prefix),
						tfjsonpath.New("url"),
						knownvalue.StringExact(edaproject1.Url),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_eda_project.test-id", configprefix.Prefix),
						tfjsonpath.New("verify_ssl"),
						knownvalue.Bool(edaproject1.VerifySsl),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_organization.test", configprefix.Prefix),
						tfjsonpath.New("eda_id"),
						fmt.Sprintf("data.%s_eda_project.test-id", configprefix.Prefix),
						tfjsonpath.New("organization_id"),
						IdCompare,
					),
				},
			},
			{
				Config: testAccEdaProjectDataSourceConfig(edaproject2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_eda_project.test-name", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(edaproject2.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_eda_project.test-name", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(edaproject2.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_eda_project.test-name", configprefix.Prefix),
						tfjsonpath.New("scm_type"),
						knownvalue.StringExact(edaproject2.ScmType),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_eda_project.test-name", configprefix.Prefix),
						tfjsonpath.New("url"),
						knownvalue.StringExact(edaproject2.Url),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_eda_project.test-name", configprefix.Prefix),
						tfjsonpath.New("verify_ssl"),
						knownvalue.Bool(edaproject2.VerifySsl),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_organization.test", configprefix.Prefix),
						tfjsonpath.New("eda_id"),
						fmt.Sprintf("data.%s_eda_project.test-name", configprefix.Prefix),
						tfjsonpath.New("organization_id"),
						IdCompare,
					),
				},
			},
		},
	})
}

func testAccEdaProjectDataSourceConfig(resource EdaProjectAPIModel) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "test" {
  name        			= "%[2]s"
}
resource "%[1]s_eda_project" "test" {
  name         			= "%[3]s"
  description  			= "%[4]s"
  scm_type     			= "%[5]s"
  url      	      		= "%[6]s"
  organization_id    	= %[1]s_organization.test.eda_id
  verify_ssl        	= %[7]v
}
data "%[1]s_eda_project" "test-id" {
  id = %[1]s_eda_project.test.id
}
data "%[1]s_eda_project" "test-name" {
  name = %[1]s_eda_project.test.name
}
  `, configprefix.Prefix, acctest.RandString(5), resource.Name, resource.Description, resource.ScmType, resource.Url, resource.VerifySsl)
}
