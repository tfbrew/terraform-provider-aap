resource "aap_organization" "example" {
  name = "Example Organization"
}
resource "aap_role_definition" "example" {
  name         = "Example Role Definition"
  description  = "Example role definition"
  content_type = "shared.organization"
  permissions  = ["shared.member_organization", "shared.view_organization"]
}
resource "aap_user" "example" {
  username   = "example"
  first_name = "First"
  last_name  = "Last"
  email      = "example@example.com"
  password   = "example-password"
}
resource "aap_role_user_assignment" "test" {
  object_id       = aap_organization.example.id
  role_definition = aap_role_definition.example.id
  user            = aap_user.example.id
}
