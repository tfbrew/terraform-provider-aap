package provider

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/tfbrew/terraform-provider-aap/internal/configprefix"
)

// Test for no input attribute supplied.
func TestAccCredentialResource_noInputAttr(t *testing.T) {

	simulatedApiValues := CredentialAPIModel{
		Name:        "test-no-input-attr-" + acctest.RandString(5),
		Description: "initial description",
	}

	t.Parallel()
	resource.Test(t, resource.TestCase{

		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCredentialNoInput(simulatedApiValues),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						fmt.Sprintf("%s_credential.test-no-input", configprefix.Prefix),
						"name",
						simulatedApiValues.Name,
					),
					resource.TestCheckResourceAttr(
						fmt.Sprintf("%s_credential.test-no-input", configprefix.Prefix),
						"description",
						simulatedApiValues.Description,
					)),
			},
			{
				PreConfig: func() {
					simulatedApiValues.Description = "updated description"
				},
				Config: testAccCredentialNoInput(simulatedApiValues),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						fmt.Sprintf("%s_credential.test-no-input", configprefix.Prefix),
						"name",
						simulatedApiValues.Name,
					),
					resource.TestCheckResourceAttr(
						fmt.Sprintf("%s_credential.test-no-input", configprefix.Prefix),
						"description",
						simulatedApiValues.Description,
					)),
			},
		},
	})

}

func TestAccCredentialResource_InputWO(t *testing.T) {

	simulatedApiValues := CredentialAPIModel{
		Name:        "test-input-wo-" + acctest.RandString(5),
		Description: "initial description",
		InputsWO: map[string]any{
			"username":  acctest.RandString(5),
			"authorize": true,
			"password":  "origpasss" + acctest.RandString(10),
		},
	}

	t.Parallel()
	resource.Test(t, resource.TestCase{

		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCredentialInputsWO(simulatedApiValues),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						fmt.Sprintf("%s_credential.test-input-wo", configprefix.Prefix),
						"name",
						simulatedApiValues.Name,
					),
					resource.TestCheckResourceAttr(
						fmt.Sprintf("%s_credential.test-input-wo", configprefix.Prefix),
						"description",
						simulatedApiValues.Description,
					)),
			},
			{
				PreConfig: func() {
					simulatedApiValues.InputsWO = map[string]any{
						"username":  acctest.RandString(5),
						"authorize": true,
						"password":  "updatedpass" + acctest.RandString(10),
					}
					simulatedApiValues.InputsWOVersion = 2
				},
				Config: testAccCredentialInputsWO(simulatedApiValues),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						fmt.Sprintf("%s_credential.test-input-wo", configprefix.Prefix),
						"name",
						simulatedApiValues.Name,
					),
					resource.TestCheckResourceAttr(
						fmt.Sprintf("%s_credential.test-input-wo", configprefix.Prefix),
						"description",
						simulatedApiValues.Description,
					)),
			},
		},
	})

}

// This tests that the input object contains only string or boolean types.
// The provided "average_cpu" being a float64 should cause an error.
func TestAccCredentialResource_inputObjTyping(t *testing.T) {

	simulatedApiValues := CredentialAPIModel{
		Name: "test-input-typing-" + acctest.RandString(5),
		Inputs: map[string]any{
			"host":        "quay.io",
			"is_active":   true,
			"average_cpu": 2.5,
		},
	}

	t.Parallel()

	resource.Test(t, resource.TestCase{

		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config:      testAccCredentialInputObjTypingConfig(simulatedApiValues),
			ExpectError: regexp.MustCompile(`inputs key 'average_cpu' has an unexpected type`),
		}},
	})

}

// This is a test that an input of type object with string and boolean values is accepted.
func TestAccCredentialResource_inputObjbasic(t *testing.T) {

	simulatedApiValues := CredentialAPIModel{
		Name: "test-input-typing-" + acctest.RandString(5),
		Inputs: map[string]any{
			"username":  acctest.RandString(5),
			"authorize": true,
			"password":  "origpass" + acctest.RandString(10),
		},
	}
	updatedPassword := "newpass" + acctest.RandString(10)
	updatePassword := func() {
		simulatedApiValues.Inputs["password"] = updatedPassword
	}

	simulatedImportApiValues := CredentialAPIModel{
		Name: "test-input-typing-" + acctest.RandString(5),
		Inputs: map[string]any{
			"username":  acctest.RandString(5),
			"authorize": true,
			"password":  "importpass" + acctest.RandString(10),
		},
	}
	t.Parallel()

	resource.Test(t, resource.TestCase{

		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: testAccCredentialInputObjTypingBasic(simulatedApiValues),
			// ExpectError: regexp.MustCompile(`inputs key 'average_cpu' has an unexpected type`),
			// Check that the resource is created successfully
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr(
					fmt.Sprintf("%s_credential.test-input-typing", configprefix.Prefix),
					"inputs.username",
					mustString(simulatedApiValues.Inputs["username"]),
				),
				resource.TestCheckResourceAttr(
					fmt.Sprintf("%s_credential.test-input-typing", configprefix.Prefix),
					"inputs.authorize",
					"true",
				),
				resource.TestCheckResourceAttr(
					fmt.Sprintf("%s_credential.test-input-typing", configprefix.Prefix),
					"inputs.password",
					mustString(simulatedApiValues.Inputs["password"]),
				),
				resource.TestCheckResourceAttr(
					fmt.Sprintf("%s_credential.test-input-typing", configprefix.Prefix),
					"name",
					simulatedApiValues.Name,
				),
			),
		},
			{
				PreConfig: updatePassword,
				Config:    testAccCredentialInputObjTypingBasic(simulatedApiValues),
				// ExpectError: regexp.MustCompile(`inputs key 'average_cpu' has an unexpected type`),
				// Check that the resource is created successfully
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						fmt.Sprintf("%s_credential.test-input-typing", configprefix.Prefix),
						"inputs.username",
						mustString(simulatedApiValues.Inputs["username"]),
					),
					resource.TestCheckResourceAttr(
						fmt.Sprintf("%s_credential.test-input-typing", configprefix.Prefix),
						"inputs.authorize",
						"true",
					),
					resource.TestCheckResourceAttr(
						fmt.Sprintf("%s_credential.test-input-typing", configprefix.Prefix),
						"inputs.password",
						mustString(simulatedApiValues.Inputs["password"]),
					),
					resource.TestCheckResourceAttr(
						fmt.Sprintf("%s_credential.test-input-typing", configprefix.Prefix),
						"name",
						simulatedApiValues.Name,
					),
				),
			},
			{
				Config: testAccCredentialInputObjTypingBasicImport(simulatedImportApiValues)},
			{
				ImportState:       true,
				ImportStateKind:   resource.ImportCommandWithID,
				ImportStateVerify: true,
				ResourceName:      fmt.Sprintf("%s_credential.test-input-import", configprefix.Prefix),
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					// The resource ID is the ID of the created credential
					rs, ok := s.RootModule().Resources[fmt.Sprintf("%s_credential.test-input-import", configprefix.Prefix)]
					if !ok {
						return "", fmt.Errorf("not found: %s_credential.test-input-import", configprefix.Prefix)
					}

					id := rs.Primary.ID
					if id == "" {
						return "", fmt.Errorf("no ID is set")
					}
					return fmt.Sprintf("%s%spassword%s%s", id, importIDSeparator, importIDSeparator, simulatedImportApiValues.Inputs["password"]), nil
				},
			},
		},
	})

}

func TestAccCredentialResource_basic(t *testing.T) {

	t.Parallel()

	// Note: The private key generation here is just for testing purposes.
	// In a real scenario, you would use a valid private key.
	// Generate a real ECDSA private key (P256 curve)
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		fmt.Println("Error generating key:", err)
		return
	}

	// Convert to PKCS8 ASN.1 DER format
	der, err := x509.MarshalECPrivateKey(privKey)
	if err != nil {
		fmt.Println("Error marshalling private key:", err)
		return
	}

	// Encode as PEM (OpenSSH-compatible format)
	pemBlock := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: der,
	}
	pemBytes := pem.EncodeToMemory(pemBlock)

	// Encode the private key in OpenSSH format
	opensshKey := string(pemBytes)

	resourceName := "test-credential-container-registry-" + acctest.RandString(5)
	resource1 := CredentialAPIModel{
		Name:        "test-credential-machine-" + acctest.RandString(5),
		Description: "test description 1",
		Inputs: map[string]any{
			"become_method":   "sudo",
			"become_password": "ASK",
			"password":        "test1234",
			"username":        "testuser",
		},
	}

	resource1inputs := mustMarshal(resource1.Inputs)

	resource2 := CredentialAPIModel{
		Name:        "test-credential-source-control-" + acctest.RandString(5),
		Description: "test description 2",
		Inputs: map[string]any{
			"ssh_key_data": opensshKey,
			"username":     "testuser",
		},
	}

	resource2inputs := mustMarshal(resource2.Inputs)

	resource3 := CredentialAPIModel{
		Name:        resourceName,
		Description: "test description 3",
		Inputs: map[string]any{
			"host":       "quay.io",
			"password":   "test1234",
			"username":   "test",
			"verify_ssl": true,
		},
	}

	resource3inputs := mustMarshal(resource3.Inputs)

	resource4 := CredentialAPIModel{
		Name:        resourceName,
		Description: "test description 4",
		Inputs: map[string]any{
			"host":       "quay.io",
			"password":   "new4567",
			"username":   "test2",
			"verify_ssl": false,
		},
	}

	resource4inputs := mustMarshal(resource4.Inputs)

	resource.Test(t, resource.TestCase{

		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCredential1Config(resource1),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential.test-machine", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource1.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential.test-machine", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(resource1.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential.test-machine", configprefix.Prefix),
						tfjsonpath.New("inputs"),
						knownvalue.StringExact(resource1inputs),
					),
				},
			},
			{
				Config: testAccCredential2Config(resource2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential.test-source-control", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource2.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential.test-source-control", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(resource2.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential.test-source-control", configprefix.Prefix),
						tfjsonpath.New("inputs"),
						knownvalue.StringExact(resource2inputs),
					),
				},
			},
			{
				Config: testAccCredential3Config(resource3),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential.test-container-registry", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource3.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential.test-container-registry", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(resource3.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential.test-container-registry", configprefix.Prefix),
						tfjsonpath.New("inputs"),
						knownvalue.StringExact(resource3inputs),
					),
				},
			},
			{
				ResourceName:            fmt.Sprintf("%s_credential.test-container-registry", configprefix.Prefix),
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"inputs"},
			},
			{
				Config: testAccCredential3Config(resource4),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential.test-container-registry", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource4.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential.test-container-registry", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(resource4.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_credential.test-container-registry", configprefix.Prefix),
						tfjsonpath.New("inputs"),
						knownvalue.StringExact(resource4inputs),
					),
				},
			},
		},
	})
}

func testAccCredential1Config(resource CredentialAPIModel) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "test-machine" {
  name        = "%[2]s"
}
data "%[1]s_credential_type" "test-machine" {
  name = "Machine"
  kind = "ssh"
}
resource "%[1]s_credential" "test-machine" {
  name            = "%[3]s"
  description	  = "%[4]s"
  organization    = %[1]s_organization.test-machine.id
  credential_type = data.%[1]s_credential_type.test-machine.id
  inputs = jsonencode(%[5]s)
}
  `, configprefix.Prefix, acctest.RandString(5), resource.Name, resource.Description, mustMarshal(resource.Inputs))
}

func testAccCredential2Config(resource CredentialAPIModel) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "test-source-control" {
  name        = "%[2]s"
}
  data "%[1]s_credential_type" "test-source-control" {
	name = "Source Control"
	kind = "scm"
  }
resource "%[1]s_credential" "test-source-control" {
  name            = "%[3]s"
  description	  = "%[4]s"
  organization    = %[1]s_organization.test-source-control.id
  credential_type = data.%[1]s_credential_type.test-source-control.id
	inputs = jsonencode(%[5]s)
}
  `, configprefix.Prefix, acctest.RandString(5), resource.Name, resource.Description, mustMarshal(resource.Inputs))
}

func testAccCredential3Config(resource CredentialAPIModel) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "test-container-registry" {
  name        = "%[2]s"
}
data "%[1]s_credential_type" "test-container-registry" {
  name = "Container Registry"
  kind = "registry"
}
resource "%[1]s_credential" "test-container-registry" {
  name            = "%[3]s"
  description	  = "%[4]s"
  organization    = %[1]s_organization.test-container-registry.id
  credential_type = data.%[1]s_credential_type.test-container-registry.id
  inputs = jsonencode(%[5]s)
}
  `, configprefix.Prefix, acctest.RandString(5), resource.Name, resource.Description, mustMarshal(resource.Inputs))
}

func testAccCredentialInputObjTypingConfig(resource CredentialAPIModel) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "test-input-typing" {
  name        = "%[2]s"
}
data "%[1]s_credential_type" "test-input-typing" {
  name = "Machine"
  kind = "ssh"
}
resource "%[1]s_credential" "test-input-typing" {
  name            = "%[3]s"
  organization    = %[1]s_organization.test-input-typing.id
  credential_type = data.%[1]s_credential_type.test-input-typing.id
  inputs = {
	     	host = "%[4]s"
		    is_active = %[5]t
		    average_cpu =%[6]f
            }
}`, configprefix.Prefix, acctest.RandString(5), resource.Name, mustString(resource.Inputs["host"]), mustBool(resource.Inputs["is_active"]), mustFloat64(resource.Inputs["average_cpu"]))
}

func testAccCredentialInputObjTypingBasic(resource CredentialAPIModel) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "test-input-typing" {
  name        = "%[2]s"
}
data "%[1]s_credential_type" "test-input-typing" {
  name = "Network"
  kind = "net"
}
resource "%[1]s_credential" "test-input-typing" {
  name            = "%[3]s"
  organization    = %[1]s_organization.test-input-typing.id
  credential_type = data.%[1]s_credential_type.test-input-typing.id
  inputs = {
	     	username = "%[4]s"
		    authorize = %[5]t
			password = "%[6]s"	
            }
}`, configprefix.Prefix, acctest.RandString(5), resource.Name, mustString(resource.Inputs["username"]), mustBool(resource.Inputs["authorize"]), mustString(resource.Inputs["password"]))
}

func testAccCredentialInputObjTypingBasicImport(resource CredentialAPIModel) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "test-input-import" {
  name        = "%[2]s"
}
data "%[1]s_credential_type" "test-input-import" {
  name = "Network"
  kind = "net"
}
resource "%[1]s_credential" "test-input-import" {
  name            = "%[3]s"
  organization    = %[1]s_organization.test-input-import.id
  credential_type = data.%[1]s_credential_type.test-input-import.id
  inputs = {
	     	username = "%[4]s"
		    authorize = %[5]t
			password = "%[6]s"	
            }
}`, configprefix.Prefix, acctest.RandString(5), resource.Name, mustString(resource.Inputs["username"]), mustBool(resource.Inputs["authorize"]), mustString(resource.Inputs["password"]))
}

func testAccCredentialNoInput(resource CredentialAPIModel) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "test-no-input" {
  name        = "%[2]s"
}
data "%[1]s_credential_type" "test-no-input" {
  name = "Network"
  kind = "net"
}
resource "%[1]s_credential" "test-no-input" {
  name            = "%[3]s"
  description     = "%[4]s"
  organization    = %[1]s_organization.test-no-input.id
  credential_type = data.%[1]s_credential_type.test-no-input.id
}`, configprefix.Prefix, acctest.RandString(5), resource.Name, resource.Description)
}

func testAccCredentialInputsWO(resource CredentialAPIModel) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "test-input-wo" {
  name        = "%[2]s"
}
data "%[1]s_credential_type" "test-input-wo" {
  name = "Network"
  kind = "net"
}
resource "%[1]s_credential" "test-input-wo" {
  name            = "%[3]s"
  description     = "%[4]s"
  organization    = %[1]s_organization.test-input-wo.id
  credential_type = data.%[1]s_credential_type.test-input-wo.id
}`, configprefix.Prefix, acctest.RandString(5), resource.Name, resource.Description)
}
