resource "azurerm_firewall" "example" {
  name                = "testfirewall"
  location            = "West Europe"
  resource_group_name = "rg"
  sku_name            = "AZFW_VNet"
  sku_tier            = "Standard"

  ip_configuration {
    name                 = "configuration"
    subnet_id            = "subnetid"
    public_ip_address_id = "publicipid"
  }
}

# Explanation: normal attributes appear first, then the ip_configuration block => valid
