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
	rName := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	idComparer := &compareTwoValuesAsStrings{}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: labelTestCaseSetup(orgName, labelName, rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_organization.%s", configprefix.Prefix, orgName),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_label.%s", configprefix.Prefix, rName),
						tfjsonpath.New("organization"),
						idComparer,
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_label.%s", configprefix.Prefix, rName),
						tfjsonpath.New("name"),
						knownvalue.StringExact(labelName),
					),
				},
			},
			{
				ResourceName:      fmt.Sprintf("%s_label.%s", configprefix.Prefix, rName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func labelTestCaseSetup(org, name string, rName string) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "%[5]s" {
	name = "%[2]s"
}

resource "%[1]s_label" "%[4]s" {
	name = "%[3]s"
	organization = %[1]s_organization.%[5]s.id
}
`, configprefix.Prefix, org, name, rName, org)
}
