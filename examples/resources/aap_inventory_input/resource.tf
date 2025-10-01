resource "aap_organization" "example" {
  name = "example"
}

resource "aap_inventory" "regular" {
  name         = "example"
  organization = aap_organization.example.id
}

resource "aap_inventory" "constructed" {
  name         = "example"
  organization = aap_organization.example.id
  kind         = "constructed"
  host_filter  = ""
}

resource "aap_inventory_input" "example" {
  constructed_inventory_id = aap_inventory.constructed.id
  input_inventory_id       = aap_inventory.regular.id
}