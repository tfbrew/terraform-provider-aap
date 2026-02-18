resource "aap_organization" "example" {
  name = "Example Organization"
}
data "aap_eda_credential_type" "example-basic-event-stream" {
  name = "Basic Event Stream"
}
resource "aap_eda_credential" "example-basic-event-stream" {
  name               = "Example Basic Event Stream Credential"
  organization_id    = aap_organization.example-basic-event-stream.eda_id
  credential_type_id = data.aap_eda_credential_type.example-basic-event-stream.id
  inputs = jsonencode({
    "username" : "testuser",
    "password" : "testpassword",
    "auth_type" : "basic",
    "http_header_key" : "Authorization",
  })
}
resource "aap_eda_event_stream" "example" {
  name                    = "Example Basic Event Stream"
  additional_data_headers = "Authorization,Content-Type"
  organization_id         = aap_organization.example-basic-event-stream.eda_id
  eda_credential_id       = aap_eda_credential.example-basic-event-stream.id
}
