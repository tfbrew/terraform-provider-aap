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

func TestAccRoleDefinitionResource(t *testing.T) {
	RoleDefinition1 := RoleDefinitionAPIModel{
		Name:        "test-group-" + acctest.RandString(5),
		Description: "Example 1",
		ContentType: "shared.organization",
		Permissions: []string{"awx.add_notificationtemplate", "awx.view_notificationtemplate", "awx.change_notificationtemplate"},
	}

	RoleDefinition2 := RoleDefinitionAPIModel{
		Name:        "test-group-" + acctest.RandString(5),
		Description: "Example 2",
		ContentType: "shared.organization",
		Permissions: []string{"awx.delete_notificationtemplate", "awx.view_notificationtemplate", "awx.change_notificationtemplate"},
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleDefinitionResourceConfig(RoleDefinition1),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_role_definition.test", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(RoleDefinition1.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_role_definition.test", configprefix.Prefix),
						tfjsonpath.New("content_type"),
						knownvalue.StringExact(RoleDefinition1.ContentType),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_role_definition.test", configprefix.Prefix),
						tfjsonpath.New("permissions"),
						knownvalue.SetExact([]knownvalue.Check{
							knownvalue.StringExact(RoleDefinition1.Permissions[0]),
							knownvalue.StringExact(RoleDefinition1.Permissions[1]),
							knownvalue.StringExact(RoleDefinition1.Permissions[2]),
						}),
					),
				},
			},
			{
				ResourceName:      fmt.Sprintf("%s_role_definition.test", configprefix.Prefix),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRoleDefinitionResourceConfig(RoleDefinition2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_role_definition.test", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(RoleDefinition2.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_role_definition.test", configprefix.Prefix),
						tfjsonpath.New("content_type"),
						knownvalue.StringExact(RoleDefinition2.ContentType),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_role_definition.test", configprefix.Prefix),
						tfjsonpath.New("permissions"),
						knownvalue.SetExact([]knownvalue.Check{
							knownvalue.StringExact(RoleDefinition2.Permissions[0]),
							knownvalue.StringExact(RoleDefinition2.Permissions[1]),
							knownvalue.StringExact(RoleDefinition2.Permissions[2]),
						}),
					),
				},
			},
		},
	})
}

func testAccRoleDefinitionResourceConfig(resource RoleDefinitionAPIModel) string {
	return fmt.Sprintf(`
resource "%[1]s_role_definition" "test" {
  name         = "%[2]s"
  description  = "Test role definition"
  content_type = "%[3]s"
  permissions   = ["%[4]s", "%[5]s", "%[6]s"]
}
`, configprefix.Prefix, resource.Name, resource.ContentType, resource.Permissions[0], resource.Permissions[1], resource.Permissions[2])
}
