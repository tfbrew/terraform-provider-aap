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

func TestAccEdaCredentialTypeResource(t *testing.T) {
	resource1 := EdaCredentialTypeAPIModel{
		Name:        "test-credential-type-" + acctest.RandString(5),
		Description: "test description 1",
		Inputs:      "{\"fields\":[{\"id\":\"username\",\"label\":\"Username\",\"type\":\"string\"},{\"id\":\"password\",\"label\":\"Password\",\"secret\":true,\"type\":\"string\"}],\"required\":[\"username\",\"password\"]}",
		Injectors:   "{\"extra_vars\":{\"ansible_password\":\"{{ password }}\",\"ansible_user\":\"{{ username }}\"}}",
	}
	resource2 := EdaCredentialTypeAPIModel{
		Name:        "test-credential-type-" + acctest.RandString(5),
		Description: "test description 2",
		Inputs:      "{\"fields\":[{\"id\":\"username\",\"label\":\"Username\",\"type\":\"string\"},{\"id\":\"password\",\"label\":\"Password\",\"secret\":true,\"type\":\"string\"}],\"required\":[\"username\",\"password\"]}",
		Injectors:   "{\"extra_vars\":{\"ansible_password\":\"{{ password }}\",\"ansible_user\":\"{{ username }}\"}}",
	}

	resource1InputsStr, ok := resource1.Inputs.(string)
	if !ok {
		t.Fatalf("Failed to convert resource1.Inputs to string")
	}
	resource1InjectorsStr, ok := resource1.Injectors.(string)
	if !ok {
		t.Fatalf("Failed to convert resource1.Injectors to string")
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEdaCredentialTypeConfig(resource1),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_eda_credential_type.test", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource1.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_eda_credential_type.test", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(resource1.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_eda_credential_type.test", configprefix.Prefix),
						tfjsonpath.New("kind"),
						knownvalue.StringExact("cloud"),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_eda_credential_type.test", configprefix.Prefix),
						tfjsonpath.New("inputs"),
						knownvalue.StringExact(resource1InputsStr),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_eda_credential_type.test", configprefix.Prefix),
						tfjsonpath.New("injectors"),
						knownvalue.StringExact(resource1InjectorsStr),
					),
				},
			},
			{
				ResourceName:      fmt.Sprintf("%s_eda_credential_type.test", configprefix.Prefix),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEdaCredentialTypeConfig(resource2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_eda_credential_type.test", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource2.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_eda_credential_type.test", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(resource2.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_eda_credential_type.test", configprefix.Prefix),
						tfjsonpath.New("kind"),
						knownvalue.StringExact("cloud"),
					),
				},
			},
		},
	})
}

func testAccEdaCredentialTypeConfig(resource EdaCredentialTypeAPIModel) string {
	return fmt.Sprintf(`
resource "%[1]s_eda_credential_type" "test" {
  name         = "%[2]s"
  description  = "%[3]s"
  inputs       = jsonencode(%[4]v)
  injectors    = jsonencode(%[5]v)
}
  `, configprefix.Prefix, resource.Name, resource.Description, resource.Inputs, resource.Injectors)
}
