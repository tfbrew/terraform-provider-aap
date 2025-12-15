package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/tfbrew/terraform-provider-aap/internal/configprefix"
)

func TestAccWorkflowJobTemplNotifSuccessResource(t *testing.T) {
	IdCompare := &compareTwoValuesAsStrings{}
	StringListCompare := &compareStringInList{}
	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowJobTemplNotifSuccess1ResourceResourceConfig(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_workflow_job_template.test", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template_notification_template_success.test", configprefix.Prefix),
						tfjsonpath.New("workflow_job_template_id"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_notification_template.test", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template_notification_template_success.test", configprefix.Prefix),
						tfjsonpath.New("notif_template_ids"),
						StringListCompare,
					),
				},
			},
			{
				ResourceName:                         fmt.Sprintf("%s_workflow_job_template_notification_template_success.test", configprefix.Prefix),
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    importStateWorkflowJobTemplateID(fmt.Sprintf("%s_workflow_job_template_notification_template_success.test", configprefix.Prefix)),
				ImportStateVerifyIdentifierAttribute: ("workflow_job_template_id"),
			},
			{
				Config: testAccWorkflowJobTemplNotifSuccess2ResourceResourceConfig(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_workflow_job_template.test", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template_notification_template_success.test", configprefix.Prefix),
						tfjsonpath.New("workflow_job_template_id"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_notification_template.test", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template_notification_template_success.test", configprefix.Prefix),
						tfjsonpath.New("notif_template_ids"),
						StringListCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_notification_template.test2", configprefix.Prefix),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template_notification_template_success.test", configprefix.Prefix),
						tfjsonpath.New("notif_template_ids"),
						StringListCompare,
					),
				},
			},
		},
	})
}

func testAccWorkflowJobTemplNotifSuccess1ResourceResourceConfig() string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "test" {
  name        = "%[2]s"
}
resource "%[1]s_workflow_job_template" "test" {
  name        = "%[2]s"
  organization 		= %[1]s_organization.test.id
}
resource "%[1]s_notification_template" "test" {
  name              = "%[2]s"
  notification_type = "slack"
  organization      = %[1]s_organization.test.id
  notification_configuration = jsonencode({
    channels  = ["#channel1", "#channel1"]
    hex_color = ""
    token     = ""
  })
  messages = jsonencode({
    error = {
      body    = ""
      message = ""
    }
    started = {
      body    = ""
      message = ""
    }
    success = {
      body    = ""
      message = "{{ job_friendly_name }} #{{ job.id }} '{{ job.name }}' {{ job.status }}: {{ url }} Custom Message"
    }
    workflow_approval = {
      approved = {
        body    = ""
        message = ""
      }
      denied = {
        body    = ""
        message = ""
      }
      running = {
        body    = ""
        message = ""
      }
      timed_out = {
        body    = ""
        message = ""
      }
    }
  })
}
resource "%[1]s_workflow_job_template_notification_template_success" "test" {
  workflow_job_template_id    = %[1]s_workflow_job_template.test.id
  notif_template_ids = [%[1]s_notification_template.test.id]
}
  `, configprefix.Prefix, acctest.RandString(5))
}

func testAccWorkflowJobTemplNotifSuccess2ResourceResourceConfig() string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "test" {
  name        = "%[2]s"
}
resource "%[1]s_workflow_job_template" "test" {
  name        = "%[2]s"
  organization 		= %[1]s_organization.test.id
}
resource "%[1]s_notification_template" "test" {
  name              = "%[2]s-2"
  notification_type = "slack"
  organization      = %[1]s_organization.test.id
  notification_configuration = jsonencode({
    channels  = ["#channel1", "#channel1"]
    hex_color = ""
    token     = ""
  })
  messages = jsonencode({
    error = {
      body    = ""
      message = ""
    }
    started = {
      body    = ""
      message = ""
    }
    success = {
      body    = ""
      message = "{{ job_friendly_name }} #{{ job.id }} '{{ job.name }}' {{ job.status }}: {{ url }} Custom Message"
    }
    workflow_approval = {
      approved = {
        body    = ""
        message = ""
      }
      denied = {
        body    = ""
        message = ""
      }
      running = {
        body    = ""
        message = ""
      }
      timed_out = {
        body    = ""
        message = ""
      }
    }
  })
}
resource "%[1]s_notification_template" "test2" {
  name              = "%[2]s-3"
  notification_type = "slack"
  organization      = %[1]s_organization.test.id
  notification_configuration = jsonencode({
    channels  = ["#channel1", "#channel1"]
    hex_color = ""
    token     = ""
  })
  messages = jsonencode({
    error = {
      body    = ""
      message = ""
    }
    started = {
      body    = ""
      message = ""
    }
    success = {
      body    = ""
      message = "{{ job_friendly_name }} #{{ job.id }} '{{ job.name }}' {{ job.status }}: {{ url }} Custom Message"
    }
    workflow_approval = {
      approved = {
        body    = ""
        message = ""
      }
      denied = {
        body    = ""
        message = ""
      }
      running = {
        body    = ""
        message = ""
      }
      timed_out = {
        body    = ""
        message = ""
      }
    }
  })
}
resource "%[1]s_workflow_job_template_notification_template_success" "test" {
  workflow_job_template_id    = %[1]s_workflow_job_template.test.id
  notif_template_ids = [%[1]s_notification_template.test.id, %[1]s_notification_template.test2.id]
}
  `, configprefix.Prefix, acctest.RandString(5))
}
