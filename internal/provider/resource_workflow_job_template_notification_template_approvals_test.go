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

func TestAccWorkflowJobTemplNotifApprovalsResource(t *testing.T) {
	rName := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha)
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
				Config: testAccWorkflowJobTemplNotifApprovals1ResourceConfig(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_workflow_job_template.%s", configprefix.Prefix, rName),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template_notification_template_approvals.%s", configprefix.Prefix, rName),
						tfjsonpath.New("workflow_job_template_id"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_notification_template.%s", configprefix.Prefix, rName+"a"),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template_notification_template_approvals.%s", configprefix.Prefix, rName),
						tfjsonpath.New("notif_template_ids"),
						StringListCompare,
					),
				},
			},
			{
				ResourceName:                         fmt.Sprintf("%s_workflow_job_template_notification_template_approvals.%s", configprefix.Prefix, rName),
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    importStateWorkflowJobTemplateID(fmt.Sprintf("%s_workflow_job_template_notification_template_approvals.%s", configprefix.Prefix, rName)),
				ImportStateVerifyIdentifierAttribute: ("workflow_job_template_id"),
			},
			{
				Config: testAccWorkflowJobTemplNotifApprovals2ResourceConfig(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_workflow_job_template.%s", configprefix.Prefix, rName),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template_notification_template_approvals.%s", configprefix.Prefix, rName),
						tfjsonpath.New("workflow_job_template_id"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_notification_template.%s", configprefix.Prefix, rName+"a"),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template_notification_template_approvals.%s", configprefix.Prefix, rName),
						tfjsonpath.New("notif_template_ids"),
						StringListCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_notification_template.%s", configprefix.Prefix, rName+"b"),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_workflow_job_template_notification_template_approvals.%s", configprefix.Prefix, rName),
						tfjsonpath.New("notif_template_ids"),
						StringListCompare,
					),
				},
			},
		},
	})
}

func testAccWorkflowJobTemplNotifApprovals1ResourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "%[3]s" {
  name        = "%[2]s"
}
resource "%[1]s_workflow_job_template" "%[3]s" {
  name        = "%[2]s"
  organization 		= %[1]s_organization.%[3]s.id
}
resource "%[1]s_notification_template" "%[4]s" {
  name              = "%[2]s"
  notification_type = "slack"
  organization      = %[1]s_organization.%[3]s.id
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
      message = ""
    }
    workflow_approval = {
      approved = {
        body    = ""
        message = "{{ job_friendly_name }} #{{ job.id }} '{{ job.name }}' {{ job.status }}: {{ url }} Custom Message"
      }
      denied = {
        body    = ""
        message = "{{ job_friendly_name }} #{{ job.id }} '{{ job.name }}' {{ job.status }}: {{ url }} Custom Message"
      }
      running = {
        body    = ""
        message = "{{ job_friendly_name }} #{{ job.id }} '{{ job.name }}' {{ job.status }}: {{ url }} Custom Message"
      }
      timed_out = {
        body    = ""
        message = "{{ job_friendly_name }} #{{ job.id }} '{{ job.name }}' {{ job.status }}: {{ url }} Custom Message"
      }
    }
  })
}
resource "%[1]s_workflow_job_template_notification_template_approvals" "%[3]s" {
  workflow_job_template_id    = %[1]s_workflow_job_template.%[3]s.id
  notif_template_ids = [%[1]s_notification_template.%[4]s.id]
}
  `, configprefix.Prefix, acctest.RandString(5), rName, rName+"a")
}

func testAccWorkflowJobTemplNotifApprovals2ResourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "%[3]s" {
  name        = "%[2]s"
}
resource "%[1]s_workflow_job_template" "%[3]s" {
  name        = "%[2]s"
  organization 		= %[1]s_organization.%[3]s.id
}
resource "%[1]s_notification_template" "%[4]s" {
  name              = "%[2]s-2"
  notification_type = "slack"
  organization      = %[1]s_organization.%[3]s.id
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
      message = ""
    }
    workflow_approval = {
      approved = {
        body    = ""
        message = "{{ job_friendly_name }} #{{ job.id }} '{{ job.name }}' {{ job.status }}: {{ url }} Custom Message"
      }
      denied = {
        body    = ""
        message = "{{ job_friendly_name }} #{{ job.id }} '{{ job.name }}' {{ job.status }}: {{ url }} Custom Message"
      }
      running = {
        body    = ""
        message = "{{ job_friendly_name }} #{{ job.id }} '{{ job.name }}' {{ job.status }}: {{ url }} Custom Message"
      }
      timed_out = {
        body    = ""
        message = "{{ job_friendly_name }} #{{ job.id }} '{{ job.name }}' {{ job.status }}: {{ url }} Custom Message"
      }
    }
  })
}
resource "%[1]s_notification_template" "%[5]s" {
  name              = "%[2]s-3"
  notification_type = "slack"
  organization      = %[1]s_organization.%[3]s.id
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
      message = ""
    }
    workflow_approval = {
      approved = {
        body    = ""
        message = "{{ job_friendly_name }} #{{ job.id }} '{{ job.name }}' {{ job.status }}: {{ url }} Custom Message"
      }
      denied = {
        body    = ""
        message = "{{ job_friendly_name }} #{{ job.id }} '{{ job.name }}' {{ job.status }}: {{ url }} Custom Message"
      }
      running = {
        body    = ""
        message = "{{ job_friendly_name }} #{{ job.id }} '{{ job.name }}' {{ job.status }}: {{ url }} Custom Message"
      }
      timed_out = {
        body    = ""
        message = "{{ job_friendly_name }} #{{ job.id }} '{{ job.name }}' {{ job.status }}: {{ url }} Custom Message"
      }
    }
  })
}
resource "%[1]s_workflow_job_template_notification_template_approvals" "%[3]s" {
  workflow_job_template_id    = %[1]s_workflow_job_template.%[3]s.id
  notif_template_ids = [%[1]s_notification_template.%[4]s.id, %[1]s_notification_template.%[5]s.id]
}
  `, configprefix.Prefix, acctest.RandString(5), rName, rName+"a", rName+"b")
}
