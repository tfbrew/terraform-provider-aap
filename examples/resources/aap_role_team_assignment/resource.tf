resource "aap_organization" "example" {
  name = "Example Organization"
}
resource "aap_role_definition" "example" {
  name         = "Example Role Definition"
  description  = "Example role definition"
  content_type = "shared.organization"
  permissions  = ["shared.member_organization", "shared.view_organization"]
}
resource "aap_team" "example" {
  name         = "Example Team"
  organization = aap_organization.example.id
  description  = "Example team description"
}
resource "aap_role_user_assignment" "test" {
  object_id       = aap_organization.example.id
  role_definition = aap_role_definition.example.id
  user            = aap_user.example.id
}
