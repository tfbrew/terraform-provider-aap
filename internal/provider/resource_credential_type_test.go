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
	rName := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
	resource1 := CredentialTypeAPIModel{
		Name:        "test-credential-type-" + acctest.RandString(5),
		Description: "test description 1",
	}
	resource2 := CredentialTypeAPIModel{
		Name:        "test-credential-type-" + acctest.RandString(5),
		Description: "test description 2",
		Kind:        "cloud",
		Inputs:      `{"fields":[{"id":"username","label":"Access Key","type":"string"}]}`,
	}
	resource2Inputs, ok := resource2.Inputs.(string)
	if !ok {
		t.Fatalf("resource2.Inputs should be a string, got %T", resource2.Inputs)
	}
	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCredentialTypeConfig1(resource1, rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential_type.%s", configprefix.Prefix, rName),
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource1.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential_type.%s", configprefix.Prefix, rName),
						tfjsonpath.New("description"),
						knownvalue.StringExact(resource1.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential_type.%s", configprefix.Prefix, rName),
						tfjsonpath.New("kind"),
						knownvalue.StringExact("cloud"), // default value is `cloud`
					),
				},
			},
			{
				ResourceName:      fmt.Sprintf("%s_credential_type.%s", configprefix.Prefix, rName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCredentialTypeConfig2(resource2, rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential_type.%s", configprefix.Prefix, rName),
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource2.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential_type.%s", configprefix.Prefix, rName),
						tfjsonpath.New("description"),
						knownvalue.StringExact(resource2.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential_type.%s", configprefix.Prefix, rName),
						tfjsonpath.New("kind"),
						knownvalue.StringExact(resource2.Kind),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential_type.%s", configprefix.Prefix, rName),
						tfjsonpath.New("inputs"),
						knownvalue.StringExact(resource2Inputs),
					),
				},
			},
		},
	})
}

func testAccCredentialTypeConfig1(resource CredentialTypeAPIModel, rName string) string {
	return fmt.Sprintf(`
resource "%[1]s_credential_type" "%[4]s" {
  name         = "%[2]s"
  description  = "%[3]s"
}
  `, configprefix.Prefix, resource.Name, resource.Description, rName)
}

func testAccCredentialTypeConfig2(resource CredentialTypeAPIModel, rName string) string {
	return fmt.Sprintf(`
resource "%[1]s_credential_type" "%[6]s" {
  name         = "%[2]s"
  description  = "%[3]s"
  kind         = "%[4]s"
  inputs	   = jsonencode(%[5]s)
}
  `, configprefix.Prefix, resource.Name, resource.Description, resource.Kind, resource.Inputs, rName)
}
