package provider

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/tfbrew/terraform-provider-aap/internal/configprefix"
)

func TestAccEdaEventStreamDataSource(t *testing.T) {
	IdCompare := &compareTwoValuesAsStrings{}
	edaeventstream1 := EdaEventStreamAPIModel{
		Name:                  "test-event-stream-" + acctest.RandString(5),
		AdditionalDataHeaders: "Authorization,Content-Type",
		TestMode:              false,
	}

	edaeventstream2 := EdaEventStreamAPIModel{
		Name: "test-event-stream-" + acctest.RandString(5),
		Uuid: uuid.New().String(),
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEdaEventStreamDataSourceConfig(edaeventstream1),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_eda_event_stream.test", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(edaeventstream1.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_eda_event_stream.test", configprefix.Prefix),
						tfjsonpath.New("additional_data_headers"),
						knownvalue.StringExact(edaeventstream1.AdditionalDataHeaders),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_eda_event_stream.test", configprefix.Prefix),
						tfjsonpath.New("test_mode"),
						knownvalue.Bool(edaeventstream1.TestMode),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_organization.test", configprefix.Prefix),
						tfjsonpath.New("eda_id"),
						fmt.Sprintf("data.%s_eda_event_stream.test", configprefix.Prefix),
						tfjsonpath.New("organization_id"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_eda_credential.test", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("data.%s_eda_event_stream.test", configprefix.Prefix),
						tfjsonpath.New("eda_credential_id"),
						IdCompare,
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_eda_event_stream.test", configprefix.Prefix),
						tfjsonpath.New("uuid"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_eda_event_stream.test", configprefix.Prefix),
						tfjsonpath.New("url"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_eda_event_stream.test", configprefix.Prefix),
						tfjsonpath.New("event_stream_type"),
						knownvalue.NotNull(),
					),
				},
			},
			{
				Config: testAccEdaEventStreamDataSourceConfig2(edaeventstream2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_eda_event_stream.test-name", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(edaeventstream2.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_eda_event_stream.test-name", configprefix.Prefix),
						tfjsonpath.New("additional_data_headers"),
						knownvalue.StringExact(edaeventstream2.AdditionalDataHeaders),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_eda_event_stream.test-name", configprefix.Prefix),
						tfjsonpath.New("uuid"),
						knownvalue.StringExact(edaeventstream2.Uuid),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_eda_event_stream.test-name", configprefix.Prefix),
						tfjsonpath.New("test_mode"),
						knownvalue.Bool(false),
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_organization.test-name", configprefix.Prefix),
						tfjsonpath.New("eda_id"),
						fmt.Sprintf("data.%s_eda_event_stream.test-name", configprefix.Prefix),
						tfjsonpath.New("organization_id"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_eda_credential.test-name", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("data.%s_eda_event_stream.test-name", configprefix.Prefix),
						tfjsonpath.New("eda_credential_id"),
						IdCompare,
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_eda_event_stream.test-name", configprefix.Prefix),
						tfjsonpath.New("url"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_eda_event_stream.test-name", configprefix.Prefix),
						tfjsonpath.New("event_stream_type"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func testAccEdaEventStreamDataSourceConfig(resource EdaEventStreamAPIModel) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "test" {
  name        			= "%[2]s"
}
data "%[1]s_eda_credential_type" "test" {
  name = "Basic Event Stream"
}
resource "%[1]s_eda_credential" "test" {
  name            = "%[2]s"
  organization_id = %[1]s_organization.test.eda_id
  credential_type_id = data.%[1]s_eda_credential_type.test.id
  inputs = jsonencode({
			"username":        "testuser",
			"password":        "testpassword",
			"auth_type":       "basic",
			"http_header_key": "Authorization",
		})
}
resource "%[1]s_eda_event_stream" "test" {
  name         			= "%[3]s"
  additional_data_headers = "%[4]s"
  test_mode				= %[5]t
  organization_id    	= %[1]s_organization.test.eda_id
  eda_credential_id		= %[1]s_eda_credential.test.id
}
data "%[1]s_eda_event_stream" "test" {
  id = %[1]s_eda_event_stream.test.id
}
  `, configprefix.Prefix, acctest.RandString(5), resource.Name, resource.AdditionalDataHeaders, resource.TestMode)
}

func testAccEdaEventStreamDataSourceConfig2(resource EdaEventStreamAPIModel) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "test-name" {
  name        			= "%[2]s"
}
data "%[1]s_eda_credential_type" "test-name" {
  name = "Basic Event Stream"
}
resource "%[1]s_eda_credential" "test-name" {
  name            = "%[2]s"
  organization_id = %[1]s_organization.test-name.eda_id
  credential_type_id = data.%[1]s_eda_credential_type.test-name.id
  inputs = jsonencode({
			"username":        "testuser",
			"password":        "testpassword",
			"auth_type":       "basic",
			"http_header_key": "Authorization",
		})
}
resource "%[1]s_eda_event_stream" "test-name" {
  name         			  = "%[3]s"
  additional_data_headers = "%[4]s"
  test_mode				  = %[5]t
  organization_id    	  = %[1]s_organization.test-name.eda_id
  eda_credential_id		  = %[1]s_eda_credential.test-name.id
  uuid					  = "%[6]s"
}
data "%[1]s_eda_event_stream" "test-name" {
  name = %[1]s_eda_event_stream.test-name.name
}
  `, configprefix.Prefix, acctest.RandString(5), resource.Name, resource.AdditionalDataHeaders, resource.TestMode, resource.Uuid)
}
