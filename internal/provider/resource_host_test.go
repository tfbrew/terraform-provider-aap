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

func TestAccHostResource(t *testing.T) {
	host1 := HostAPIModel{
		Name:        "test-host-" + acctest.RandString(5),
		Description: "Example with jsonencoded variables for localhost",
		Variables:   "{\"foo\":\"bar\"}",
		Enabled:     true,
	}

	host2 := HostAPIModel{
		Name:        "test-host-" + acctest.RandString(5),
		Description: "Updated example with different variables",
		Enabled:     false,
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccHostResourceConfig1(host1),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_host.test", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(host1.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_host.test", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(host1.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_host.test", configprefix.Prefix),
						tfjsonpath.New("variables"),
						knownvalue.StringExact(host1.Variables),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_host.test", configprefix.Prefix),
						tfjsonpath.New("enabled"),
						knownvalue.Bool(host1.Enabled),
					),
				},
			},
			{
				ResourceName:      fmt.Sprintf("%s_host.test", configprefix.Prefix),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccHostResourceConfig2(host2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_host.test", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(host2.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_host.test", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(host2.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_host.test", configprefix.Prefix),
						tfjsonpath.New("variables"),
						knownvalue.Null(),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_host.test", configprefix.Prefix),
						tfjsonpath.New("enabled"),
						knownvalue.Bool(host2.Enabled),
					),
				},
			},
		},
	})
}

func testAccHostResourceConfig1(resource HostAPIModel) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "test" {
  name        = "test-organization-%[2]s"
  description = "test"
}
resource "%[1]s_inventory" "test" {
  name         = "test-inventory-%[2]s"
  description  = "test"
  organization = %[1]s_organization.test.id
}
resource "%[1]s_host" "test" {
  name        = "%[3]s"
  description = "%[4]s"
  inventory   = %[1]s_inventory.test.id
  variables   = jsonencode(%[5]s)
  enabled     = %[6]v
}
  `, configprefix.Prefix, acctest.RandString(5), resource.Name, resource.Description, resource.Variables, resource.Enabled)
}

func testAccHostResourceConfig2(resource HostAPIModel) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "test" {
  name        = "test-organization-%[2]s"
  description = "test"
}
resource "%[1]s_inventory" "test" {
  name         = "test-inventory-%[2]s"
  description  = "test"
  organization = %[1]s_organization.test.id
}
resource "%[1]s_host" "test" {
  name        = "%[3]s"
  description = "%[4]s"
  inventory   = %[1]s_inventory.test.id
  enabled     = %[5]v
}
  `, configprefix.Prefix, acctest.RandString(5), resource.Name, resource.Description, resource.Enabled)
}
