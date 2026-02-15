resource "azurerm_container_app" "example" {
  name                = "example"
  resource_group_name = "rg-example"

  identity {
    type = "SystemAssigned"
  }

  tags = {
    environment = "dev"
  }

  lifecycle {}

  depends_on = []
}
