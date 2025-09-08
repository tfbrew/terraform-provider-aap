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

func TestAccCredentialTypeResource(t *testing.T) {
	resource1 := CredentialTypeAPIModel{
		Name:        "test-credential-type-" + acctest.RandString(5),
		Description: "test description 1",
		Kind:        "cloud",
	}
	resource2 := CredentialTypeAPIModel{
		Name:        "test-credential-type-" + acctest.RandString(5),
		Description: "test description 2",
		Kind:        "cloud",
	}
	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCredentialTypeConfig(resource1),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential_type.test", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource1.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential_type.test", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(resource1.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential_type.test", configprefix.Prefix),
						tfjsonpath.New("kind"),
						knownvalue.StringExact(resource1.Kind),
					),
				},
			},
			{
				ResourceName:      fmt.Sprintf("%s_credential_type.test", configprefix.Prefix),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCredentialTypeConfig(resource2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential_type.test", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource2.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential_type.test", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(resource2.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential_type.test", configprefix.Prefix),
						tfjsonpath.New("kind"),
						knownvalue.StringExact(resource1.Kind),
					),
				},
			},
		},
	})
}

func testAccCredentialTypeConfig(resource CredentialTypeAPIModel) string {
	return fmt.Sprintf(`
resource "%[1]s_credential_type" "test" {
  name         = "%[2]s"
  description  = "%[3]s"
}
  `, configprefix.Prefix, resource.Name, resource.Description)
}
