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

func TestAccJobTemplNotifErrResource(t *testing.T) {
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
				Config: testAccJobTemplNotifErr1ResourceConfig(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_job_template.%s", configprefix.Prefix, rName),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_job_template_notification_template_error.%s", configprefix.Prefix, rName),
						tfjsonpath.New("job_template_id"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_notification_template.%s", configprefix.Prefix, rName+"a"),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_job_template_notification_template_error.%s", configprefix.Prefix, rName),
						tfjsonpath.New("notif_template_ids"),
						StringListCompare,
					),
				},
			},
			{
				ResourceName:                         fmt.Sprintf("%s_job_template_notification_template_error.%s", configprefix.Prefix, rName),
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    importStateJobTemplateID(fmt.Sprintf("%s_job_template_notification_template_error.%s", configprefix.Prefix, rName)),
				ImportStateVerifyIdentifierAttribute: ("job_template_id"),
			},
			{
				Config: testAccJobTemplNotifErr2ResourceConfig(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_job_template.%s", configprefix.Prefix, rName),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_job_template_notification_template_error.%s", configprefix.Prefix, rName),
						tfjsonpath.New("job_template_id"),
						IdCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_notification_template.%s", configprefix.Prefix, rName+"b"),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_job_template_notification_template_error.%s", configprefix.Prefix, rName),
						tfjsonpath.New("notif_template_ids"),
						StringListCompare,
					),
					statecheck.CompareValuePairs(
						fmt.Sprintf("%s_notification_template.%s", configprefix.Prefix, rName+"c"),
						tfjsonpath.New("id"),
						fmt.Sprintf("%s_job_template_notification_template_error.%s", configprefix.Prefix, rName),
						tfjsonpath.New("notif_template_ids"),
						StringListCompare,
					),
				},
			},
		},
	})
}

func testAccJobTemplNotifErr1ResourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "%[3]s" {
  name        = "%[2]s"
}

resource "%[1]s_inventory" "%[3]s" {
  name         = "%[2]s"
  organization = %[1]s_organization.%[3]s.id
}

resource "%[1]s_project" "%[3]s" {
  name         		= "%[2]s"
  organization 		= %[1]s_organization.%[3]s.id
  scm_type     		= "git"
  scm_url      		= "git@github.com:user/repo.git"
  allow_override 	= true
}

resource "%[1]s_job_template" "%[3]s" {
  name        = "%[2]s"
  job_type    = "run"
  inventory   = %[1]s_inventory.%[3]s.id
  project     = %[1]s_project.%[3]s.id
  playbook    = "test.yml"
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
      message = "{{ job_friendly_name }} #{{ job.id }} '{{ job.name }}' {{ job.status }}: {{ url }} Custom Message"
    }
    success = {
      body    = ""
      message = ""
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
resource "%[1]s_job_template_notification_template_error" "%[3]s" {
  job_template_id    = %[1]s_job_template.%[3]s.id
  notif_template_ids = [%[1]s_notification_template.%[4]s.id]
}
  `, configprefix.Prefix, acctest.RandString(5), rName, rName+"a")
}

func testAccJobTemplNotifErr2ResourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "%[1]s_organization" "%[3]s" {
  name        = "%[2]s"
}

resource "%[1]s_inventory" "%[3]s" {
  name         = "%[2]s"
  organization = %[1]s_organization.%[3]s.id
}

resource "%[1]s_project" "%[3]s" {
  name         		= "%[2]s"
  organization 		= %[1]s_organization.%[3]s.id
  scm_type     		= "git"
  scm_url      		= "git@github.com:user/repo.git"
  allow_override 	= true
}

resource "%[1]s_job_template" "%[3]s" {
  name        = "%[2]s"
  job_type    = "run"
  inventory   = %[1]s_inventory.%[3]s.id
  project     = %[1]s_project.%[3]s.id
  playbook    = "test.yml"
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
      message = "{{ job_friendly_name }} #{{ job.id }} '{{ job.name }}' {{ job.status }}: {{ url }} Custom Message"
    }
    success = {
      body    = ""
      message = ""
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
      message = "{{ job_friendly_name }} #{{ job.id }} '{{ job.name }}' {{ job.status }}: {{ url }} Custom Message"
    }
    success = {
      body    = ""
      message = ""
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
resource "%[1]s_job_template_notification_template_error" "%[3]s" {
  job_template_id    = %[1]s_job_template.%[3]s.id
  notif_template_ids = [%[1]s_notification_template.%[4]s.id, %[1]s_notification_template.%[5]s.id]
}
  `, configprefix.Prefix, acctest.RandString(5), rName, rName+"b", rName+"c")
}
