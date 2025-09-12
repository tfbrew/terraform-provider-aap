resource "aap_organization" "example" {
  name        = "example"
  description = "example"
}

resource "aap_inventory" "example" {
  name         = "example"
  description  = "example"
  organization = aap_organization.example.id
}

resource "aap_project" "example" {
  name         = "example"
  description  = "example"
  organization = aap_organization.example.id
  scm_type     = "git"
  scm_url      = "<SCM_URL>"
}

resource "aap_job_template" "example" {
  job_type  = "run"
  name      = "test"
  inventory = aap_inventory.example.id
  project   = aap_project.example.id
  playbook  = "test.yml"
}
