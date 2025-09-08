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

func TestAccCredentialTypeDataSource(t *testing.T) {
	resource1 := CredentialTypeAPIModel{
		Name:        "test-credential-type-" + acctest.RandString(5),
		Description: "test description",
		Kind:        "cloud",
	}
	resource2 := CredentialTypeAPIModel{
		Name:        "test-credential-type-" + acctest.RandString(5),
		Description: "test description",
		Kind:        "cloud",
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
				Config: testAccCredentialTypeDataSourceIDConfig(resource1),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_credential_type.test-id", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource1.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_credential_type.test-id", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(resource1.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_credential_type.test-id", configprefix.Prefix),
						tfjsonpath.New("kind"),
						knownvalue.StringExact(resource1.Kind),
					),
				},
			},
			// Read by name testing
			{
				Config: testAccCredentialTypeDataSourceNameConfig(resource2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_credential_type.test-name", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource2.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_credential_type.test-name", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(resource2.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_credential_type.test-name", configprefix.Prefix),
						tfjsonpath.New("kind"),
						knownvalue.StringExact(resource2.Kind),
					),
				},
			},
		},
	})
}

func testAccCredentialTypeDataSourceIDConfig(resource CredentialTypeAPIModel) string {
	return fmt.Sprintf(`
resource "%[1]s_credential_type" "test-id" {
  name         = "%[2]s"
  description  = "%[3]s"
}
data "%[1]s_credential_type" "test-id" {
  id = %[1]s_credential_type.test-id.id
}
  `, configprefix.Prefix, resource.Name, resource.Description)
}

func testAccCredentialTypeDataSourceNameConfig(resource CredentialTypeAPIModel) string {
	return fmt.Sprintf(`
resource "%[1]s_credential_type" "test-name" {
  name         = "%[2]s"
  description  = "%[3]s"
}
data "%[1]s_credential_type" "test-name" {
  name = %[1]s_credential_type.test-name.name
  kind = %[1]s_credential_type.test-name.kind
}
  `, configprefix.Prefix, resource.Name, resource.Description)
}
