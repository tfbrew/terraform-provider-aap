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

func TestAccRoleDefinitionDataSource(t *testing.T) {
	roleDef := RoleDefinitionAPIModel{
		Name:        "test-roledef-" + acctest.RandString(5),
		Description: "Test role definition datasource",
		ContentType: "shared.organization",
		Permissions: []string{"awx.add_notificationtemplate", "awx.view_notificationtemplate"},
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleDefinitionDataSourceConfig(roleDef),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_role_definition.test-id", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(roleDef.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_role_definition.test-id", configprefix.Prefix),
						tfjsonpath.New("content_type"),
						knownvalue.StringExact(roleDef.ContentType),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_role_definition.test-id", configprefix.Prefix),
						tfjsonpath.New("permissions"),
						knownvalue.SetExact([]knownvalue.Check{
							knownvalue.StringExact(roleDef.Permissions[0]),
							knownvalue.StringExact(roleDef.Permissions[1]),
						}),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_role_definition.test-name", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(roleDef.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_role_definition.test-name", configprefix.Prefix),
						tfjsonpath.New("content_type"),
						knownvalue.StringExact(roleDef.ContentType),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_role_definition.test-name", configprefix.Prefix),
						tfjsonpath.New("permissions"),
						knownvalue.SetExact([]knownvalue.Check{
							knownvalue.StringExact(roleDef.Permissions[0]),
							knownvalue.StringExact(roleDef.Permissions[1]),
						}),
					),
				},
			},
		},
	})
}

func testAccRoleDefinitionDataSourceConfig(resource RoleDefinitionAPIModel) string {
	return fmt.Sprintf(`
resource "%[1]s_role_definition" "test" {
  name         = "%[2]s"
  description  = "%[3]s"
  content_type = "%[4]s"
  permissions  = ["%[5]s", "%[6]s"]
}

data "%[1]s_role_definition" "test-id" {
  id = %[1]s_role_definition.test.id
}

data "%[1]s_role_definition" "test-name" {
  name = %[1]s_role_definition.test.name
}
`, configprefix.Prefix, resource.Name, resource.Description, resource.ContentType, resource.Permissions[0], resource.Permissions[1])
}
