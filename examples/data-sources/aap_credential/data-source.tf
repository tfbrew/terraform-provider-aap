data "aap_credential" "example" {
  id = "1"
}

data "aap_credential" "example-name" {
  name            = "Demo Credential"
  credential_type = 1
  organization    = 1
}
