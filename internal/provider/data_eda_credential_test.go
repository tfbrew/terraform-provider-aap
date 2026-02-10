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

func TestAccEdaCredentialDataSource(t *testing.T) {

	resource1 := EdaCredentialAPIModel{
		Name:        "test-credential-basic-event-stream-" + acctest.RandString(5),
		Description: "test description 2",
		Inputs: map[string]any{
			"username":        "testuser",
			"password":        "testpassword",
			"auth_type":       "basic",
			"http_header_key": "Authorization",
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
				Config: testAccEdaCredentialDataSourceConfig(resource1),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_eda_credential.test", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource1.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_eda_credential.test", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(resource1.Description),
					),
				},
			},
		},
	})
}

func testAccEdaCredentialDataSourceConfig(resource EdaCredentialAPIModel) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "test" {
  name        = "%[2]s"
}
data "%[1]s_eda_credential_type" "test" {
  name = "Basic Event Stream"
}
resource "%[1]s_eda_credential" "test" {
  name            = "%[3]s"
  description	  = "%[4]s"
  organization_id = %[1]s_organization.test.eda_id
  credential_type_id = data.%[1]s_eda_credential_type.test.id
  inputs = jsonencode(%[5]s)
}
data "%[1]s_eda_credential" "test" {
  id = %[1]s_eda_credential.test.id
}
  `, configprefix.Prefix, acctest.RandString(5), resource.Name, resource.Description, mustMarshal(resource.Inputs))
}
