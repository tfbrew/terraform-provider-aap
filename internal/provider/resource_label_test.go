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

func TestAccLabel_basic(t *testing.T) {
	orgName := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	labelName := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	idComparer := &compareTwoValuesAsStrings{}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: labelTestCaseSetup(orgName, labelName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_organization.testorg", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_label.test", configprefix.Prefix),
						tfjsonpath.New("organization"),
						idComparer,
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_label.test", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(labelName),
					),
				},
			},
			{
				ResourceName:      fmt.Sprintf("%s_label.test", configprefix.Prefix),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func labelTestCaseSetup(org, name string) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "testorg" {
	name = "%[2]s"
}

resource "%[1]s_label" "test" {
	name = "%[3]s"
	organization = %[1]s_organization.testorg.id
}
`, configprefix.Prefix, org, name)
}
