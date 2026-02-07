resource "aap_eda_credential_type" "example" {
  name        = "Example EDA Credential Type"
  description = "Example credential type for authentication"
  inputs = jsonencode(
    {
      "fields" : [
        {
          "id" : "username",
          "label" : "Username",
          "type" : "string"
        },
        {
          "id" : "password",
          "label" : "Password",
          "secret" : true,
          "type" : "string"
        }
      ],
      "required" : ["username", "password"]
    }
  )
  injectors = jsonencode(
    {
      "extra_vars" : {
        "ansible_user" : "{{ username }}",
        "ansible_password" : "{{ password }}"
      }
    }
  )
}
