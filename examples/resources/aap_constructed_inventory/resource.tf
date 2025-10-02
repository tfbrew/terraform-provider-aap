resource "aap_organization" "example" {
  name        = "Example Organization"
  description = "An example organization"
}

resource "aap_inventory" "example" {
  name         = "Example Inventory"
  description  = "An example inventory"
  organization = aap_organization.example.id
}

resource "aap_constructed_inventory" "example" {
  name         = "Example Constructed Inventory"
  description  = "An example constructed inventory"
  organization = aap_organization.example.id
  limit        = "dev,&no_mgmt"
  source_vars  = <<-EOT
---
plugin: ansible.builtin.constructed
strict: true
groups:
  dev: az_env_tag == "Development"
  no_mgmt: not ("mgmt" in inventory_hostname)
EOT
}

resource "aap_inventory_input" "example" {
  constructed_inventory_id = aap_constructed_inventory.example.id
  input_inventory_id       = aap_inventory.example.id
}  