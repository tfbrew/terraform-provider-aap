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

func TestAccTeamDataSource(t *testing.T) {
	IdCompare := &compareTwoValuesAsStrings{}
	teamName := "test-team-" + acctest.RandString(5)
	teamDesc := "Test team description"
	orgName := "test-org-" + acctest.RandString(5)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Lookup by id
			{
				Config: testAccTeamDataSourceIdConfig(teamName, teamDesc, orgName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_team.test-id", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(teamName),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_team.test-id", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(teamDesc),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_organization.test", configprefix.Prefix),
						tfjsonpath.New("aap25_gateway_id"),
						fmt.Sprintf("data.%s_team.test-id", configprefix.Prefix),
						tfjsonpath.New("organization"),
						IdCompare,
					),
				},
			},
			// Lookup by name
			{
				Config: testAccTeamDataSourceNameConfig(teamName, teamDesc, orgName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_team.test-name", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(teamName),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_team.test-name", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(teamDesc),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_organization.test", configprefix.Prefix),
						tfjsonpath.New("aap25_gateway_id"),
						fmt.Sprintf("data.%s_team.test-name", configprefix.Prefix),
						tfjsonpath.New("organization"),
						IdCompare,
					),
				},
			},
		},
	})
}

func testAccTeamDataSourceIdConfig(teamName, teamDesc, orgName string) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "test" {
  name = "%[2]s"
}

resource "%[1]s_team" "test" {
  name         = "%[3]s"
  organization = %[1]s_organization.test.aap25_gateway_id
  description  = "%[4]s"
}

data "%[1]s_team" "test-id" {
  id = %[1]s_team.test.id
}
`, configprefix.Prefix, orgName, teamName, teamDesc)
}

func testAccTeamDataSourceNameConfig(teamName, teamDesc, orgName string) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "test" {
  name = "%[2]s"
}

resource "%[1]s_team" "test" {
  name         = "%[3]s"
  organization = %[1]s_organization.test.aap25_gateway_id
  description  = "%[4]s"
}

data "%[1]s_team" "test-name" {
  name = %[1]s_team.test.name
}
`, configprefix.Prefix, orgName, teamName, teamDesc)
}
