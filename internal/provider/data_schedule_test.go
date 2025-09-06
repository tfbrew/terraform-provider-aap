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

func TestAccScheduleDataSource(t *testing.T) {
	schedule := ScheduleAPIModel{
		Name:               "test-schedule-" + acctest.RandString(5),
		Description:        "Initial test schedule",
		Rrule:              "DTSTART;TZID=UTC:20250301T120000 RRULE:FREQ=DAILY;INTERVAL=1",
		UnifiedJobTemplate: 1,
		Enabled:            true,
	}
	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0), // built-in check from tfversion package
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read by ID testing
			{
				Config: testAccScheduleDataSourceConfig(schedule),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_schedule.test", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(schedule.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_schedule.test", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(schedule.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_schedule.test", configprefix.Prefix),
						tfjsonpath.New("rrule"),
						knownvalue.StringExact(schedule.Rrule),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_schedule.test", configprefix.Prefix),
						tfjsonpath.New("unified_job_template"),
						knownvalue.Int32Exact(int32(schedule.UnifiedJobTemplate)),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.%s_schedule.test", configprefix.Prefix),
						tfjsonpath.New("enabled"),
						knownvalue.Bool(schedule.Enabled),
					),
				},
			},
		},
	})
}

func testAccScheduleDataSourceConfig(resource ScheduleAPIModel) string {
	return fmt.Sprintf(`
resource "%[1]s_schedule" "test" {
  name        			= "%[2]s"
  description 			= "%[3]s"
  rrule       			= "%[4]s"
  unified_job_template 	= %[5]d
  enabled     			= %[6]t
}
data "%[1]s_schedule" "test" {
  id = %[1]s_schedule.test.id
}
`, configprefix.Prefix, resource.Name, resource.Description, resource.Rrule, resource.UnifiedJobTemplate, resource.Enabled)
}
