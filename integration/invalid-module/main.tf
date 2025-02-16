  
resource "azurerm_resource_group" "example" {

  tags     = var.tags

  location = var.location
  name     = var.resource_group_name

}

module "dummy" {
  source     = "./dummy_module"
  depends_on = [azurerm_resource_group.example]
}
