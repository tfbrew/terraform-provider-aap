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

func TestAccGenericEndpointResource(t *testing.T) {

	apiPath1 := "authenticators"
	apiEndpoint1 := "gateway"
	dataJson1 := fmt.Sprintf(`{"auto_migrate_users_to":null,"configuration":{},"create_objects":true,"enabled":false,"name":"test-%s","order":%d,"remove_users":true,"type":"ansible_base.authentication.authenticator_plugins.local"}`, acctest.RandString(5), acctest.RandIntRange(1000, 9999))
	dataJson1_1 := fmt.Sprintf(`{"auto_migrate_users_to":null,"configuration":{},"create_objects":false,"enabled":false,"name":"test-%s","order":%d,"remove_users":true,"type":"ansible_base.authentication.authenticator_plugins.local"}`, acctest.RandString(5), acctest.RandIntRange(1000, 9999))

	apiPath2 := "authenticator_maps"
	apiEndpoint2 := "gateway"
	dataJson2 := fmt.Sprintf(`{"authenticator":1,"map_type":"allow","name":"test-%s","order":%d,"organization":"","revoke":false,"role":"","team":"","triggers":{"always":{}}}`, acctest.RandString(5), acctest.RandIntRange(1000, 9999))

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGenericEndpointConfig(1, apiPath1, apiEndpoint1, dataJson1),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_generic_endpoint.test-1", configprefix.Prefix),
						tfjsonpath.New("data_json"),
						knownvalue.StringExact(dataJson1),
					),
				},
			},
			{
				Config: testAccGenericEndpointConfig(1, apiPath1, apiEndpoint1, dataJson1_1),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_generic_endpoint.test-1", configprefix.Prefix),
						tfjsonpath.New("data_json"),
						knownvalue.StringExact(dataJson1_1),
					),
				},
			},
			{
				Config: testAccGenericEndpointConfig(2, apiPath2, apiEndpoint2, dataJson2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_generic_endpoint.test-2", configprefix.Prefix),
						tfjsonpath.New("data_json"),
						knownvalue.StringExact(dataJson2),
					),
				},
			},
		},
	})
}

func testAccGenericEndpointConfig(count int, apiPath string, apiEndpoint string, dataJson string) string {
	return fmt.Sprintf(`
resource "%[1]s_generic_endpoint" "test-%[2]d" {
  api_path       = "%[3]s"
  api_endpoint   = "%[4]s"
  data_json      = jsonencode(%[5]s)
}
	`, configprefix.Prefix, count, apiPath, apiEndpoint, dataJson)
}
