resource "aap_organization" "example" {
  name = "example"
}

resource "aap_inventory" "regular" {
  name         = "example"
  organization = aap_organization.example.id
}

resource "aap_constructed_inventory" "constructed" {
  name         = "example"
  organization = aap_organization.example.id
}

resource "aap_inventory_input" "example" {
  constructed_inventory_id = aap_constructed_inventory.constructed.id
  input_inventory_id       = aap_inventory.regular.id
}