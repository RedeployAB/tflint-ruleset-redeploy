provider "azurerm" {
  features {}
}

provider "azurerm" {
  features {}
  alias = "secondary"
}

resource "azurerm_resource_group" "example" {

  tags = var.tags

  location = var.location
  name     = var.resource_group_name

}

module "dummy" {
  source     = "./dummy_module"
  depends_on = [azurerm_resource_group.example]
}

resource "azurerm_subnet" "from_length" {
  count = length(["a", "b"])

  lifecycle {
    prevent_destroy = false
  }
}
