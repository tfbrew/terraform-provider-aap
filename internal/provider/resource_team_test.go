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

func TestAccTeamResource(t *testing.T) {
	rName := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	IdCompare := &compareTwoValuesAsStrings{}
	teamName := "test-team-" + acctest.RandString(5)
	teamDesc := "Test team description"
	teamDesc2 := "Test team description"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTeamResourceConfig(teamName, teamDesc, rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_team.%s", configprefix.Prefix, rName),
						tfjsonpath.New("name"),
						knownvalue.StringExact(teamName),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_team.%s", configprefix.Prefix, rName),
						tfjsonpath.New("description"),
						knownvalue.StringExact(teamDesc),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_organization.%s", configprefix.Prefix, rName),
						tfjsonpath.New("gateway_id"),
						fmt.Sprintf("%s_team.%s", configprefix.Prefix, rName),
						tfjsonpath.New("organization"),
						IdCompare,
					),
				},
			},
			{
				ResourceName:      fmt.Sprintf("%s_team.%s", configprefix.Prefix, rName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTeamResourceConfig(teamName, teamDesc2, rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_team.%s", configprefix.Prefix, rName),
						tfjsonpath.New("name"),
						knownvalue.StringExact(teamName),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_team.%s", configprefix.Prefix, rName),
						tfjsonpath.New("description"),
						knownvalue.StringExact(teamDesc2),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_organization.%s", configprefix.Prefix, rName),
						tfjsonpath.New("gateway_id"),
						fmt.Sprintf("%s_team.%s", configprefix.Prefix, rName),
						tfjsonpath.New("organization"),
						IdCompare,
					),
				},
			},
		},
	})
}

func testAccTeamResourceConfig(teamName, teamDesc string, rName string) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "%[5]s" {
  name = "%[2]s"
}

resource "%[1]s_team" "%[5]s" {
  name         = "%[3]s"
  organization = %[1]s_organization.%[5]s.gateway_id
  description  = "%[4]s"
}
`, configprefix.Prefix, acctest.RandString(5), teamName, teamDesc, rName)
}
