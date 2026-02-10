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

func TestAccEdaCredentialResource(t *testing.T) {
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

	resource1inputs := mustMarshal(resource1.Inputs)

	resourceName := "test-credential-container-registry-" + acctest.RandString(5)

	resource2 := EdaCredentialAPIModel{
		Name:        resourceName,
		Description: "test description 3",
		Inputs: map[string]any{
			"host":       "quay.io",
			"password":   "test1234",
			"username":   "test",
			"verify_ssl": true,
		},
	}

	resource2inputs := mustMarshal(resource2.Inputs)

	resource3 := EdaCredentialAPIModel{
		Name:        resourceName,
		Description: "test description 4",
		Inputs: map[string]any{
			"host":       "quay.io",
			"password":   "new4567",
			"username":   "test2",
			"verify_ssl": false,
		},
	}

	resource3inputs := mustMarshal(resource3.Inputs)

	resource.Test(t, resource.TestCase{

		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEdaCredential1Config(resource1),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_eda_credential.test-basic-event-stream", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource1.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_eda_credential.test-basic-event-stream", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(resource1.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_eda_credential.test-basic-event-stream", configprefix.Prefix),
						tfjsonpath.New("inputs"),
						knownvalue.StringExact(resource1inputs),
					),
				},
			},
			{
				Config: testAccEdaCredential2Config(resource2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_eda_credential.test-container-registry", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource2.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_eda_credential.test-container-registry", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(resource2.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_eda_credential.test-container-registry", configprefix.Prefix),
						tfjsonpath.New("inputs"),
						knownvalue.StringExact(resource2inputs),
					),
				},
			},
			{
				ResourceName:            fmt.Sprintf("%s_eda_credential.test-container-registry", configprefix.Prefix),
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"inputs"},
			},
			{
				Config: testAccEdaCredential2Config(resource3),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_eda_credential.test-container-registry", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource3.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_eda_credential.test-container-registry", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(resource3.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_eda_credential.test-container-registry", configprefix.Prefix),
						tfjsonpath.New("inputs"),
						knownvalue.StringExact(resource3inputs),
					),
				},
			},
		},
	})
}

func testAccEdaCredential1Config(resource EdaCredentialAPIModel) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "test-basic-event-stream" {
  name        = "%[2]s"
}
data "%[1]s_eda_credential_type" "test-basic-event-stream" {
  name = "Basic Event Stream"
}
resource "%[1]s_eda_credential" "test-basic-event-stream" {
  name            = "%[3]s"
  description	  = "%[4]s"
  organization_id = %[1]s_organization.test-basic-event-stream.eda_id
  credential_type_id = data.%[1]s_eda_credential_type.test-basic-event-stream.id
  inputs = jsonencode(%[5]s)
}
  `, configprefix.Prefix, acctest.RandString(5), resource.Name, resource.Description, mustMarshal(resource.Inputs))
}

func testAccEdaCredential2Config(resource EdaCredentialAPIModel) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "test-container-registry" {
  name        = "%[2]s"
}
data "%[1]s_eda_credential_type" "test-container-registry" {
  name = "Container Registry"
}
resource "%[1]s_eda_credential" "test-container-registry" {
  name            = "%[3]s"
  description	  = "%[4]s"
  organization_id = %[1]s_organization.test-container-registry.eda_id
  credential_type_id = data.%[1]s_eda_credential_type.test-container-registry.id
  inputs = jsonencode(%[5]s)
}
  `, configprefix.Prefix, acctest.RandString(5), resource.Name, resource.Description, mustMarshal(resource.Inputs))
}
