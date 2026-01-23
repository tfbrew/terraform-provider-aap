terraform {
  required_providers {
    aap = {
      source = "tfbrew/aap"
    }
  }
}

# OAuth token authentication
provider "aap" {
  endpoint = "https://aap.example.com"
  token    = "oauth_token_here"
}

# API token authentication
provider "aap" {
  endpoint  = "https://aap.example.com"
  api_token = "api_token_here"
}

# Username/password authentication
provider "aap" {
  endpoint = "http://aap.example.com"
  username = "admin"
  password = "password"
}

# With API retry configuration
provider "aap" {
  endpoint = "http://aap.example.com"
  token    = "mysecrettoken"
  api_retry = {
    api_retry_count         = 1
    api_retry_delay_seconds = 2
  }
}
