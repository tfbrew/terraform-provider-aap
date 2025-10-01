package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/tfbrew/terraform-provider-aap/internal/configprefix"
)

func TestAccInventoryInputResource_basic(t *testing.T) {

	IdCompare := &compareTwoValuesAsStrings{}

	orgName, constructedInvName, inputInvName := acctest.RandString(5), acctest.RandString(5), acctest.RandString(5)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInventoryInputConfig(orgName, constructedInvName, inputInvName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_inventory_input.test", configprefix.Prefix),
						tfjsonpath.New("constructed_inventory_id"),
						fmt.Sprintf("%s_constructed_inventory.constructed_test", configprefix.Prefix),
						tfjsonpath.New("id"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_inventory_input.test", configprefix.Prefix),
						tfjsonpath.New("input_inventory_id"),
						fmt.Sprintf("%s_inventory.regular_test", configprefix.Prefix),
						tfjsonpath.New("id"),
						IdCompare,
					),
				},
			},
		},
	})
}

func TestAccInventoryInputResource_import(t *testing.T) {
	orgName, constructedInvName, inputInvName := acctest.RandString(5), acctest.RandString(5), acctest.RandString(5)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInventoryInputConfig(orgName, constructedInvName, inputInvName),
			},
			{
				ResourceName:                         fmt.Sprintf("%s_inventory_input.test", configprefix.Prefix),
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "constructed_inventory_id",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					// Find the resource in the state
					rs, ok := s.RootModule().Resources[fmt.Sprintf("%s_constructed_inventory.constructed_test", configprefix.Prefix)]
					if !ok {
						return "", fmt.Errorf("constructed inventory not found")
					}
					id := rs.Primary.ID
					if id == "" {
						return "", fmt.Errorf("constructed inventory has no ID")
					}

					rs, ok = s.RootModule().Resources[fmt.Sprintf("%s_inventory.regular_test", configprefix.Prefix)]
					if !ok {
						return "", fmt.Errorf("input inventory not found")
					}
					inputID := rs.Primary.ID
					if inputID == "" {
						return "", fmt.Errorf("input inventory has no ID")
					}

					return fmt.Sprintf("%s,%s", id, inputID), nil
				},
			},
		},
	})
}

func TestAccInventoryInputResource_import_invalid(t *testing.T) {
	orgName, constructedInvName, inputInvName := acctest.RandString(5), acctest.RandString(5), acctest.RandString(5)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInventoryInputConfig(orgName, constructedInvName, inputInvName),
			},
			{
				ResourceName:      fmt.Sprintf("%s_inventory_input.test", configprefix.Prefix),
				ImportState:       true,
				ImportStateVerify: false,
				ImportStateIdFunc: func(_ *terraform.State) (string, error) {
					return "badvalue", nil
				},
				ExpectError: regexp.MustCompile("Invalid import id string"),
			},
		},
	})
}

func testAccInventoryInputConfig(orgName, constructedInvName, inputInvName string) string {
	return fmt.Sprintf(`
	    resource "%[1]s_organization" "test" {
			name = "%[2]s"
		}
		
		resource "%[1]s_inventory" "regular_test" {
			name =  "%[4]s"
			organization = %[1]s_organization.test.id
		}

		resource "%[1]s_constructed_inventory" "constructed_test" {
			name = "%[3]s"
			organization = %[1]s_organization.test.id
		}

		resource "%[1]s_inventory_input" "test" {
			constructed_inventory_id = %[1]s_constructed_inventory.constructed_test.id
			input_inventory_id       = %[1]s_inventory.regular_test.id
		}`,
		configprefix.Prefix, orgName, constructedInvName, inputInvName,
	)
}

func testAccInventoryInputReplaceConfig(orgName, constructedInvName, inputInvName, inputReplaceInvName string) string {
	return fmt.Sprintf(`
	    resource "%[1]s_organization" "test" {
			name = "%[2]s"
		}

		resource "%[1]s_inventory" "regular_test" {
			name =  "%[4]s"
			organization = %[1]s_organization.test.id
		}
		
		resource "%[1]s_inventory" "regular_test_2" {
			name =  "%[5]s"
			organization = %[1]s_organization.test.id
		}

		resource "%[1]s_constructed_inventory" "constructed_test" {
			name = "%[3]s"
			organization = %[1]s_organization.test.id

		}

		resource "%[1]s_inventory_input" "test" {
			constructed_inventory_id = %[1]s_constructed_inventory.constructed_test.id
			input_inventory_id       = %[1]s_inventory.regular_test_2.id
		}`,
		configprefix.Prefix, orgName, constructedInvName, inputInvName, inputReplaceInvName,
	)
}

// Verify that the Destroy to recreate is working.
func TestAccInventoryInputResource_replace(t *testing.T) {

	orgName, constructedInvName, inputInvName := acctest.RandString(5), acctest.RandString(5), acctest.RandString(5)
	secondInvName := acctest.RandString(5)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInventoryInputConfig(orgName, constructedInvName, inputInvName),
			},
			{
				Config: testAccInventoryInputReplaceConfig(orgName, constructedInvName, inputInvName, secondInvName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							fmt.Sprintf("%s_inventory_input.test", configprefix.Prefix),
							plancheck.ResourceActionDestroyBeforeCreate,
						),
					},
				},
			},
		},
	})
}

func TestAccInventoryInputResource_delete(t *testing.T) {
	orgName, constructedInvName, inputInvName := acctest.RandString(5), acctest.RandString(5), acctest.RandString(5)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInventoryInputConfig(orgName, constructedInvName, inputInvName),
			},
			{
				Config:             "{}",
				ExpectNonEmptyPlan: false,
				Check: func(s *terraform.State) error {
					if _, ok := s.RootModule().Resources[fmt.Sprintf("%s_inventory_input.test", configprefix.Prefix)]; ok {
						return fmt.Errorf("resource still exists in state")
					}
					return nil
				},
			},
		},
	})
}
