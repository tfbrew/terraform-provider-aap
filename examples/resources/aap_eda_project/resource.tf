resource "aap_organization" "example" {
  name = "Example Organization"
}
resource "aap_eda_project" "example" {
  name            = "Example EDA Project"
  description     = "Example Description"
  scm_type        = "git"
  url             = "https://github.com/tfbrew/example-repo"
  organization_id = aap_organization.test.eda_id
}
