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

func TestAccProjectDataSource(t *testing.T) {
	IdCompare := &compareTwoValuesAsStrings{}
	project1 := ProjectAPIModel{
		Name:        "test-project-" + acctest.RandString(5),
		Description: "Test git project",
		ScmType:     "git",
		ScmUrl:      "https://github.com/example/repo.git",
		Timeout:     1,
	}
	project2 := ProjectAPIModel{
		Name:        "test-project-" + acctest.RandString(5),
		Description: "svn project",
		ScmType:     "svn",
		ScmUrl:      "svn://bad_ip/test_repo",
		Timeout:     1,
	}
	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0), // built-in check from tfversion package
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read by ID testing
			{
				Config: testAccProjectDataSourceIDConfig(project1),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_project.test-id", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(project1.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_project.test-id", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(project1.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_project.test-id", configprefix.Prefix),
						tfjsonpath.New("scm_type"),
						knownvalue.StringExact(project1.ScmType),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_project.test-id", configprefix.Prefix),
						tfjsonpath.New("scm_url"),
						knownvalue.StringExact(project1.ScmUrl),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_project.test-id", configprefix.Prefix),
						tfjsonpath.New("timeout"),
						knownvalue.Int32Exact(int32(project1.Timeout)),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_organization.test-id", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("data.%s_project.test-id", configprefix.Prefix),
						tfjsonpath.New("organization"),
						IdCompare,
					),
				},
			},
			// Read by name testing
			{
				Config: testAccProjectDataSourceNameConfig(project2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_project.test-name", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(project2.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_project.test-name", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(project2.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_project.test-name", configprefix.Prefix),
						tfjsonpath.New("scm_type"),
						knownvalue.StringExact(project2.ScmType),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_project.test-name", configprefix.Prefix),
						tfjsonpath.New("scm_url"),
						knownvalue.StringExact(project2.ScmUrl),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_project.test-name", configprefix.Prefix),
						tfjsonpath.New("timeout"),
						knownvalue.Int32Exact(int32(project2.Timeout)),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_organization.test-name", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("data.%s_project.test-name", configprefix.Prefix),
						tfjsonpath.New("organization"),
						IdCompare,
					),
				},
			},
		},
	})
}

func testAccProjectDataSourceIDConfig(resource ProjectAPIModel) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "test-id" {
  name        			= "%[2]s"
}
resource "%[1]s_project" "test-id" {
  name         	= "%[3]s"
  description  	= "%[4]s"
  scm_type     	= "%[5]s"
  scm_url      	= "%[6]s"
  organization 	= %[1]s_organization.test-id.id
  timeout		= %[7]d
}
data "%[1]s_project" "test-id" {
  id = %[1]s_project.test-id.id
}
`, configprefix.Prefix, acctest.RandString(5), resource.Name, resource.Description, resource.ScmType, resource.ScmUrl, resource.Timeout)
}

func testAccProjectDataSourceNameConfig(resource ProjectAPIModel) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "test-name" {
  name        			= "%[2]s"
}
resource "%[1]s_project" "test-name" {
  name         	= "%[3]s"
  description  	= "%[4]s"
  scm_type     	= "%[5]s"
  scm_url      	= "%[6]s"
  organization 	= %[1]s_organization.test-name.id
  timeout		= %[7]d
}
data "%[1]s_project" "test-name" {
  name = %[1]s_project.test-name.name
}
`, configprefix.Prefix, acctest.RandString(5), resource.Name, resource.Description, resource.ScmType, resource.ScmUrl, resource.Timeout)
}
