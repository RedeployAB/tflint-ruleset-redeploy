resource "azurerm_container_app" "example" {
  lifecycle {
    ignore_changes = [tags]
  }

  name                = "example"
  resource_group_name = "rg-example"

  identity {
    type = "SystemAssigned"
  }

  tags = {
    environment = "dev"
  }
}
