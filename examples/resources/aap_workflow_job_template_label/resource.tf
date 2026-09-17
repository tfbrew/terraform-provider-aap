resource "aap_organization" "example" {
  name = "example-organization"
}

resource "aap_workflow_job_template" "example" {
  name         = "example-workflow-job-template"
  organization = aap_organization.example.id
}

resource "aap_label" "example-1" {
  name         = "example-label-1"
  organization = aap_organization.example.id
}

resource "aap_label" "example-2" {
  name         = "example-label-2"
  organization = aap_organization.example.id
}

resource "aap_workflow_job_template_label" "example" {
  label_ids                = [aap_label.example-1.id, aap_label.example-2.id]
  workflow_job_template_id = aap_workflow_job_template.example.id
}
