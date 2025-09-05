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

func TestAccCredentialDataSource(t *testing.T) {

	resource1 := CredentialAPIModel{
		Name:        "test-credential-" + acctest.RandString(5),
		Description: "test description",
		Inputs: map[string]any{
			"become_method":   "sudo",
			"become_password": "ASK",
			"password":        "test1234",
			"username":        "tester",
		},
	}
	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0), // built-in check from tfversion package
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCredentialDataSourceConfig(resource1),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_credential.test", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource1.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_credential.test", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(resource1.Description),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("data.%s_credential_type.test", configprefix.Prefix),
						tfjsonpath.New("kind"),
						fmt.Sprintf("%s_credential.test", configprefix.Prefix),
						tfjsonpath.New("kind"),
						compare.ValuesSame(),
					),
				},
			},
		},
	})
}

func testAccCredentialDataSourceConfig(resource CredentialAPIModel) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "test" {
  name        = "%[2]s"
}
data "%[1]s_credential_type" "test" {
  name = "Machine"
  kind = "ssh"
}
resource "%[1]s_credential" "test" {
  name            = "%[3]s"
  description	  = "%[4]s"
  organization    = %[1]s_organization.test.id
  credential_type = data.%[1]s_credential_type.test.id
  inputs = jsonencode(%[5]s)
}
data "%[1]s_credential" "test" {
  id = %[1]s_credential.test.id
}
  `, configprefix.Prefix, acctest.RandString(5), resource.Name, resource.Description, mustMarshal(resource.Inputs))
}
