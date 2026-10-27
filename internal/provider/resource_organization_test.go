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

func TestAccOrganizationResource(t *testing.T) {
	resource1 := OrganizationAPIModel{
		Name:        "test-organization-" + acctest.RandString(5),
		Description: "test description 1",
		MaxHosts:    100,
		// AAP-AWX-DIFF-BEGIN: aap sets opa_query_path; awx sets default_environment.
		OpaQueryPath: "aap_policy_examples/allowed_false",
		// AAP-AWX-DIFF-END
	}
	resource2 := OrganizationAPIModel{
		Name:        "test-organization-" + acctest.RandString(5),
		Description: "test description 1",
		MaxHosts:    200,
		// AAP-AWX-DIFF-BEGIN: aap clears opa_query_path; awx changes default_environment.
		// AAP-AWX-DIFF-END
	}
	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0), // built-in check from tfversion package
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationResourceConfig(resource1),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_organization.test", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource1.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_organization.test", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(resource1.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_organization.test", configprefix.Prefix),
						tfjsonpath.New("max_hosts"),
						knownvalue.Int32Exact(int32(resource1.MaxHosts)),
					),
					// AAP-AWX-DIFF-BEGIN: aap also checks the EDA organization ID and opa_query_path; awx checks default_environment.
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_organization.test", configprefix.Prefix),
						tfjsonpath.New("eda_id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_organization.test", configprefix.Prefix),
						tfjsonpath.New("opa_query_path"),
						knownvalue.StringExact(resource1.OpaQueryPath),
					),
					// AAP-AWX-DIFF-END
				},
			},
			// ImportState testing
			{
				ResourceName:      fmt.Sprintf("%s_organization.test", configprefix.Prefix),
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccOrganizationResourceConfig(resource2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_organization.test", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(resource2.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_organization.test", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(resource2.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_organization.test", configprefix.Prefix),
						tfjsonpath.New("max_hosts"),
						knownvalue.Int32Exact(int32(resource2.MaxHosts)),
					),
					// AAP-AWX-DIFF-BEGIN: aap also checks the EDA organization ID and opa_query_path; awx checks default_environment.
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_organization.test", configprefix.Prefix),
						tfjsonpath.New("eda_id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_organization.test", configprefix.Prefix),
						tfjsonpath.New("opa_query_path"),
						knownvalue.StringExact(""),
					),
					// AAP-AWX-DIFF-END
				},
			},
		},
	})
}

func testAccOrganizationResourceConfig(resource OrganizationAPIModel) string {
	// AAP-AWX-DIFF-BEGIN: aap sets opa_query_path; awx sets default_environment.
	return fmt.Sprintf(`
resource "%[1]s_organization" "test" {
  name           = "%[2]s"
  description    = "%[3]s"
  max_hosts      = %[4]d
  opa_query_path = "%[5]s"
}
  `, configprefix.Prefix, resource.Name, resource.Description, resource.MaxHosts, resource.OpaQueryPath)
	// AAP-AWX-DIFF-END
}
