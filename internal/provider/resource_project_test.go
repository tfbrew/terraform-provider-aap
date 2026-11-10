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

func TestAccProjectResource(t *testing.T) {
	rName := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	rName2 := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	rName3 := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	rName4 := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	IdCompare := &compareTwoValuesAsStrings{}
	project1 := ProjectAPIModel{
		Name:           "test-project-" + acctest.RandString(5),
		Description:    "Initial test git project",
		ScmType:        "git",
		ScmUrl:         "https://github.com/example/repo.git",
		Organization:   1,
		ScmUpdOnLaunch: false,
		Timeout:        1,
	}

	project2 := ProjectAPIModel{
		Name:           "test-project-" + acctest.RandString(5),
		Description:    "Updated test git project",
		ScmType:        "git",
		ScmUrl:         "https://github.com/example/updated-repo.git",
		Organization:   1,
		ScmUpdOnLaunch: true,
		Timeout:        1,
	}

	project3 := ProjectAPIModel{
		Name:           "test-project-" + acctest.RandString(5),
		Description:    "svn project",
		ScmType:        "svn",
		ScmUrl:         "svn://bad_ip/test_repo",
		Organization:   1,
		ScmUpdOnLaunch: false,
		Timeout:        1,
	}

	project4 := ProjectAPIModel{
		Name:                  "test-project-" + acctest.RandString(5),
		Description:           "archive project",
		ScmType:               "archive",
		ScmUrl:                "https://github.com/user/repo",
		Organization:          1,
		ScmUpdOnLaunch:        true,
		ScmUpdateCacheTimeout: 60,
		Timeout:               1,
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectResourceConfig(project1, rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_project.%s", configprefix.Prefix, rName),
						tfjsonpath.New("name"),
						knownvalue.StringExact(project1.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_project.%s", configprefix.Prefix, rName),
						tfjsonpath.New("description"),
						knownvalue.StringExact(project1.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_project.%s", configprefix.Prefix, rName),
						tfjsonpath.New("scm_type"),
						knownvalue.StringExact(project1.ScmType),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_project.%s", configprefix.Prefix, rName),
						tfjsonpath.New("scm_url"),
						knownvalue.StringExact(project1.ScmUrl),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_project.%s", configprefix.Prefix, rName),
						tfjsonpath.New("scm_update_on_launch"),
						knownvalue.Bool(project1.ScmUpdOnLaunch),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_project.%s", configprefix.Prefix, rName),
						tfjsonpath.New("timeout"),
						knownvalue.Int32Exact(int32(project1.Timeout)),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_organization.%s", configprefix.Prefix, rName),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_project.%s", configprefix.Prefix, rName),
						tfjsonpath.New("organization"),
						IdCompare,
					),
				},
			},
			{
				ResourceName:      fmt.Sprintf("%s_project.%s", configprefix.Prefix, rName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectResourceConfig(project2, rName2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_project.%s", configprefix.Prefix, rName2),
						tfjsonpath.New("name"),
						knownvalue.StringExact(project2.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_project.%s", configprefix.Prefix, rName2),
						tfjsonpath.New("description"),
						knownvalue.StringExact(project2.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_project.%s", configprefix.Prefix, rName2),
						tfjsonpath.New("scm_type"),
						knownvalue.StringExact(project2.ScmType),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_project.%s", configprefix.Prefix, rName2),
						tfjsonpath.New("scm_url"),
						knownvalue.StringExact(project2.ScmUrl),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_project.%s", configprefix.Prefix, rName2),
						tfjsonpath.New("scm_update_on_launch"),
						knownvalue.Bool(project2.ScmUpdOnLaunch),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_project.%s", configprefix.Prefix, rName2),
						tfjsonpath.New("timeout"),
						knownvalue.Int32Exact(int32(project2.Timeout)),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_organization.%s", configprefix.Prefix, rName2),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_project.%s", configprefix.Prefix, rName2),
						tfjsonpath.New("organization"),
						IdCompare,
					),
				},
			},
			{
				Config: testAccProjectResource3Config(project3, rName3),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_project.%s", configprefix.Prefix, rName3),
						tfjsonpath.New("name"),
						knownvalue.StringExact(project3.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_project.%s", configprefix.Prefix, rName3),
						tfjsonpath.New("description"),
						knownvalue.StringExact(project3.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_project.%s", configprefix.Prefix, rName3),
						tfjsonpath.New("scm_type"),
						knownvalue.StringExact(project3.ScmType),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_project.%s", configprefix.Prefix, rName3),
						tfjsonpath.New("scm_url"),
						knownvalue.StringExact(project3.ScmUrl),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_project.%s", configprefix.Prefix, rName3),
						tfjsonpath.New("scm_update_on_launch"),
						knownvalue.Bool(project3.ScmUpdOnLaunch),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_project.%s", configprefix.Prefix, rName3),
						tfjsonpath.New("timeout"),
						knownvalue.Int32Exact(int32(project3.Timeout)),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_organization.%s", configprefix.Prefix, rName3),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_project.%s", configprefix.Prefix, rName3),
						tfjsonpath.New("organization"),
						IdCompare,
					),
				},
			},
			{
				Config: testAccProjectResource4Config(project4, rName4),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_project.%s", configprefix.Prefix, rName4),
						tfjsonpath.New("name"),
						knownvalue.StringExact(project4.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_project.%s", configprefix.Prefix, rName4),
						tfjsonpath.New("description"),
						knownvalue.StringExact(project4.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_project.%s", configprefix.Prefix, rName4),
						tfjsonpath.New("scm_type"),
						knownvalue.StringExact(project4.ScmType),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_project.%s", configprefix.Prefix, rName4),
						tfjsonpath.New("scm_url"),
						knownvalue.StringExact(project4.ScmUrl),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_project.%s", configprefix.Prefix, rName4),
						tfjsonpath.New("scm_update_on_launch"),
						knownvalue.Bool(project4.ScmUpdOnLaunch),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_project.%s", configprefix.Prefix, rName4),
						tfjsonpath.New("timeout"),
						knownvalue.Int32Exact(int32(project4.Timeout)),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_organization.%s", configprefix.Prefix, rName4),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_project.%s", configprefix.Prefix, rName4),
						tfjsonpath.New("organization"),
						IdCompare,
					),
				},
			},
		},
	})
}

func testAccProjectResourceConfig(resource ProjectAPIModel, rName string) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "%[9]s" {
  name        			= "%[2]s"
}
resource "%[1]s_project" "%[9]s" {
  name         			= "%[3]s"
  description  			= "%[4]s"
  scm_type     			= "%[5]s"
  scm_url      			= "%[6]s"
  organization 			= %[1]s_organization.%[9]s.id
  scm_update_on_launch 	= %[7]v
  timeout				= %[8]d
}
  `, configprefix.Prefix, acctest.RandString(5), resource.Name, resource.Description, resource.ScmType, resource.ScmUrl, resource.ScmUpdOnLaunch, resource.Timeout, rName)
}

func testAccProjectResource3Config(resource ProjectAPIModel, rName string) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "%[9]s" {
  name        			= "%[2]s"
}
resource "%[1]s_project" "%[9]s" {
  name         			= "%[3]s"
  description  			= "%[4]s"
  scm_type     			= "%[5]s"
  scm_url      			= "%[6]s"
  organization 			= %[1]s_organization.%[9]s.id
  scm_update_on_launch 	= %[7]v
  timeout				= %[8]d
}
  `, configprefix.Prefix, acctest.RandString(5), resource.Name, resource.Description, resource.ScmType, resource.ScmUrl, resource.ScmUpdOnLaunch, resource.Timeout, rName)
}

func testAccProjectResource4Config(resource ProjectAPIModel, rName string) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "%[10]s" {
  name        			= "%[2]s"
}
resource "%[1]s_project" "%[10]s" {
  name         			   = "%[3]s"
  description  			   = "%[4]s"
  scm_type     			   = "%[5]s"
  scm_url      			   = "%[6]s"
  organization 			   = %[1]s_organization.%[10]s.id
  scm_update_on_launch 	   = %[7]v
  scm_update_cache_timeout = %[8]d
  timeout				   = %[9]d
}
  `, configprefix.Prefix, acctest.RandString(5), resource.Name, resource.Description, resource.ScmType, resource.ScmUrl, resource.ScmUpdOnLaunch, resource.ScmUpdateCacheTimeout, resource.Timeout, rName)
}
