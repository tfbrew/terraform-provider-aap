terraform {
  required_providers {
    aap = {
      source = "tfbrew/aap"
    }
  }
}

provider "aap" {
  endpoint = "https://tower.example.com"
  token    = "token"
}

provider "aap" {
  endpoint = "http://tower.example.com"
  username = "admin"
  password = "password"
}

provider "aap" {
  endpoint = "http://aap.example.com"
  token    = "mysecrettoken"
  api_retry = {
    api_retry_count         = 1
    api_retry_delay_seconds = 2
  }
}