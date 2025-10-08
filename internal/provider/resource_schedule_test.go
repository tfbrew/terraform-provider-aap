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

func TestAccScheduleResource(t *testing.T) {
	schedule1 := ScheduleAPIModel{
		Name:               "test-schedule-" + acctest.RandString(5),
		Description:        "Initial test schedule",
		Rrule:              "DTSTART;TZID=UTC:20250301T120000 RRULE:FREQ=DAILY;INTERVAL=1",
		UnifiedJobTemplate: 1,
		Enabled:            true,
	}

	schedule2 := ScheduleAPIModel{
		Name:               "test-schedule-" + acctest.RandString(5),
		Description:        "Updated test schedule",
		Rrule:              "DTSTART;TZID=UTC:20250301T140000 RRULE:FREQ=WEEKLY;INTERVAL=1",
		UnifiedJobTemplate: 1,
		Enabled:            false,
	}

	schedule3 := ScheduleAPIModel{
		Name:        "test-schedule-" + acctest.RandString(5),
		Description: "Updated test schedule",
		Rrule:       "DTSTART;TZID=UTC:20250301T140000 RRULE:FREQ=WEEKLY;INTERVAL=1",
		Enabled:     false,
		Verbosity:   2,
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleResourceConfig(schedule1),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_schedule.test", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(schedule1.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_schedule.test", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(schedule1.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_schedule.test", configprefix.Prefix),
						tfjsonpath.New("rrule"),
						knownvalue.StringExact(schedule1.Rrule),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_schedule.test", configprefix.Prefix),
						tfjsonpath.New("unified_job_template"),
						knownvalue.Int32Exact(int32(schedule1.UnifiedJobTemplate)),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_schedule.test", configprefix.Prefix),
						tfjsonpath.New("enabled"),
						knownvalue.Bool(schedule1.Enabled),
					),
				},
			},
			{
				ResourceName:      fmt.Sprintf("%s_schedule.test", configprefix.Prefix),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleResourceConfig(schedule2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_schedule.test", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(schedule2.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_schedule.test", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(schedule2.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_schedule.test", configprefix.Prefix),
						tfjsonpath.New("rrule"),
						knownvalue.StringExact(schedule2.Rrule),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_schedule.test", configprefix.Prefix),
						tfjsonpath.New("unified_job_template"),
						knownvalue.Int32Exact(int32(schedule2.UnifiedJobTemplate)),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_schedule.test", configprefix.Prefix),
						tfjsonpath.New("enabled"),
						knownvalue.Bool(schedule2.Enabled),
					),
				},
			},
			// The step below is to verify that it handles at least one optional field that can be used in prompting
			{
				Config: testAccScheduleResourceOptionalConfig(schedule3),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_schedule.test", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(schedule3.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_schedule.test", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(schedule3.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_schedule.test", configprefix.Prefix),
						tfjsonpath.New("rrule"),
						knownvalue.StringExact(schedule3.Rrule),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_schedule.test", configprefix.Prefix),
						tfjsonpath.New("enabled"),
						knownvalue.Bool(schedule3.Enabled),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_schedule.test", configprefix.Prefix),
						tfjsonpath.New("verbosity"),
						knownvalue.Int32Exact(int32(schedule3.Verbosity)),
					),
				},
			},
		},
	})
}

func testAccScheduleResourceConfig(resource ScheduleAPIModel) string {
	return fmt.Sprintf(`
resource "%[1]s_schedule" "test" {
  name        			= "%[2]s"
  description 			= "%[3]s"
  rrule       			= "%[4]s"
  unified_job_template 	= %[5]d
  enabled     			= %[6]t
}
  `, configprefix.Prefix, resource.Name, resource.Description, resource.Rrule, resource.UnifiedJobTemplate, resource.Enabled)
}

func testAccScheduleResourceOptionalConfig(resource ScheduleAPIModel) string {
	return fmt.Sprintf(`

data "%[1]s_organization" "default" {
	name = "Default"
}

resource "%[1]s_project" "temp_project" {
  name = "Temp Project for Schedule JT %[2]s"
  organization = data.%[1]s_organization.default.id
  scm_type     = "git"
  scm_url      = "git@github.com:user/repo.git"
  allow_override 	= true
}

resource "%[1]s_inventory" "test" {
  name         = "%[2]s"
  organization = data.%[1]s_organization.default.id
}

resource "%[1]s_job_template" "test_launch" {
	name = "test-launch-%[2]s"
	description = "Test job template for launching from schedule"
	job_type = "check"
	inventory = %[1]s_inventory.test.id
	project = %[1]s_project.temp_project.id
	playbook = "test.yml"
	ask_verbosity_on_launch = true
}

resource "%[1]s_schedule" "test" {
  name        			= "%[2]s"
  description 			= "%[3]s"
  rrule       			= "%[4]s"
  unified_job_template 	= %[1]s_job_template.test_launch.id
  enabled     			= %[5]t
  verbosity             = %[6]d
}
  `, configprefix.Prefix, resource.Name, resource.Description, resource.Rrule, resource.Enabled, resource.Verbosity)
}

func TestAccSchedule2Resource(t *testing.T) {
	schedule1 := ScheduleAPIModel{
		Name:        "test-schedule-" + acctest.RandString(5),
		Description: "Initial test schedule",
		Rrule:       "DTSTART;TZID=UTC:20250301T120000 RRULE:FREQ=DAILY;INTERVAL=1",
		Enabled:     true,
		ExtraData: map[string]any{
			"edString": "value1",
			"edInt":    3,
			"edFloat":  4.3,
			"edList":   []string{"one", "two", "three"},
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleResourceExtraDataConfig(schedule1),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_schedule.test", configprefix.Prefix),
						tfjsonpath.New("name"),
						knownvalue.StringExact(schedule1.Name),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_schedule.test", configprefix.Prefix),
						tfjsonpath.New("description"),
						knownvalue.StringExact(schedule1.Description),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_schedule.test", configprefix.Prefix),
						tfjsonpath.New("rrule"),
						knownvalue.StringExact(schedule1.Rrule),
					),
					statecheck.ExpectKnownValue(
						fmt.Sprintf("%s_schedule.test", configprefix.Prefix),
						tfjsonpath.New("enabled"),
						knownvalue.Bool(schedule1.Enabled),
					),
				},
			},
		},
	})
}

func testAccScheduleResourceExtraDataConfig(resource ScheduleAPIModel) string {
	return fmt.Sprintf(`

data "%[1]s_organization" "default" {
	name = "Default"
}

resource "%[1]s_project" "temp_project2" {
  name = "Temp Project for Schedule JT %[2]s"
  organization = data.%[1]s_organization.default.id
  scm_type     = "git"
  scm_url      = "git@github.com:user/repo.git"
  allow_override 	= true
}

resource "%[1]s_inventory" "test2" {
  name         = "%[2]s"
  organization = data.%[1]s_organization.default.id
}

resource "%[1]s_job_template" "test_launch2" {
	name = "test-launch-%[2]s"
	description = "Test job template for launching from schedule"
	job_type = "check"
	inventory = %[1]s_inventory.test2.id
	project = %[1]s_project.temp_project2.id
	playbook = "test.yml"
	ask_verbosity_on_launch = true
	survey_enabled = true
	ask_variables_on_launch = true
}

resource "%[1]s_schedule" "test" {
  name        			= "%[2]s"
  description 			= "%[3]s"
  rrule       			= "%[4]s"
  unified_job_template 	= %[1]s_job_template.test_launch2.id
  enabled     			= %[5]t
  verbosity             = %[6]d
  extra_data = {
    edString = "value1"
	edInt = 3
	edFloat = 4.3
	edList = tolist(["one", "two", "three"])
  }
}
  `, configprefix.Prefix, resource.Name, resource.Description, resource.Rrule, resource.Enabled, resource.Verbosity)
}
