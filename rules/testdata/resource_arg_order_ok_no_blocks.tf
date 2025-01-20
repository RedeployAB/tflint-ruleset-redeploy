resource "azurerm_resource_group" "example" {
  name     = "example"
  location = "West Europe"
}

# Only non-block arguments => valid
