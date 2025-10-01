package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/tfbrew/terraform-provider-aap/internal/configprefix"
)

func TestAccConstructedInventory_basic(t *testing.T) {
	name := "test-constructed-inv-" + acctest.RandString(5)
	description := "desc-" + acctest.RandString(5)
	orgName := "org-" + acctest.RandString(5)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConstructedInventoryConfig(name, description, orgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						fmt.Sprintf("%s_constructed_inventory.test", configprefix.Prefix),
						"name", name,
					),
					resource.TestCheckResourceAttr(
						fmt.Sprintf("%s_constructed_inventory.test", configprefix.Prefix),
						"description", description,
					),
				),
			},
		},
	})
}

func TestAccConstructedInventory_update(t *testing.T) {
	name := "test-constructed-inv-" + acctest.RandString(5)
	description := "desc-" + acctest.RandString(5)
	newDescription := "desc-updated-" + acctest.RandString(5)
	orgName := "org-" + acctest.RandString(5)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConstructedInventoryConfig(name, description, orgName),
			},
			{
				Config: testAccConstructedInventoryConfig(name, newDescription, orgName),
				Check: resource.TestCheckResourceAttr(
					fmt.Sprintf("%s_constructed_inventory.test", configprefix.Prefix),
					"description", newDescription,
				),
			},
		},
	})
}

func TestAccConstructedInventory_import(t *testing.T) {
	name := "test-constructed-inv-" + acctest.RandString(5)
	description := "desc-" + acctest.RandString(5)
	orgName := "org-" + acctest.RandString(5)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConstructedInventoryConfig(name, description, orgName),
			},
			{
				ResourceName:      fmt.Sprintf("%s_constructed_inventory.test", configprefix.Prefix),
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[fmt.Sprintf("%s_constructed_inventory.test", configprefix.Prefix)]
					if !ok {
						return "", fmt.Errorf("constructed inventory not found")
					}
					id := rs.Primary.ID
					if id == "" {
						return "", fmt.Errorf("constructed inventory has no ID")
					}
					return id, nil
				},
			},
		},
	})
}

func TestAccConstructedInventory_delete(t *testing.T) {
	name := "test-constructed-inv-" + acctest.RandString(5)
	description := "desc-" + acctest.RandString(5)
	orgName := "org-" + acctest.RandString(5)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConstructedInventoryConfig(name, description, orgName),
			},
			{
				Config:             "{}",
				ExpectNonEmptyPlan: false,
				Check: func(s *terraform.State) error {
					if _, ok := s.RootModule().Resources[fmt.Sprintf("%s_constructed_inventory.test", configprefix.Prefix)]; ok {
						return fmt.Errorf("resource still exists in state")
					}
					return nil
				},
			},
		},
	})
}

func testAccConstructedInventoryConfig(name, description, orgName string) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "test" {
  name        = "%[4]s"
}

resource "%[1]s_constructed_inventory" "test" {
  name        = "%[2]s"
  description = "%[3]s"
  organization = %[1]s_organization.test.id
}
`, configprefix.Prefix, name, description, orgName)
}
