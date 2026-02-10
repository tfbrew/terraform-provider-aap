resource "aap_organization" "example" {
  name        = "tf_example"
  description = "example"
}

// Example basic event stream

// Inputs options for basic event stream credentials:
// username, password, auth_type, http_header_key
// Setting a value to "ASK" is equal to choosing "Prompt at Launch"

data "aap_eda_credential_type" "basic-event-stream" {
  name = "Basic Event Stream"
}

resource "aap_eda_credential" "example-basic-event-stream" {
  name            = "example_basic_event_stream"
  organization    = aap_organization.example.eda_id
  credential_type = data.aap_credential_type.basic-event-stream.id
  inputs = jsonencode({
    "password" : "test1234", // code should not contain secrets, example only
    "username" : "aap",
    "auth_type" : "basic",
    "http_header_key" : "Authorization"
  })
}

// Example container registry credential

// Inputs options for source control credentials:
// host, username, password, verify_ssl
// Setting a value to "ASK" is equal to choosing "Prompt at Launch"

data "aap_eda_credential_type" "container-registry" {
  name = "Container Registry"
}

resource "aap_eda_credential" "example-container-registry" {
  name            = "example_container_registry"
  organization    = aap_organization.example.eda_id
  credential_type = data.aap_eda_credential_type.container-registry.id
  inputs = jsonencode({
    "host" : "quay.io",
    "password" : "test1234", // code should not contain secrets, example only
    "username" : "test",
    "verify_ssl" : true
  })
}
