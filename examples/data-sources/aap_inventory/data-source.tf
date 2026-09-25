data "aap_inventory" "example" {
  id = "1"
}

data "aap_inventory" "example-name" {
  name         = "Demo Inventory"
  organization = 1
}
