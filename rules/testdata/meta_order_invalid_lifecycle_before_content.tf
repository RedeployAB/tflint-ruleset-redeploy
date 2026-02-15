resource "azurerm_container_app" "example" {
  lifecycle {}

  name                = "example"
  resource_group_name = "rg-example"

  identity {
    type = "SystemAssigned"
  }

  tags = {
    environment = "dev"
  }
}
