resource "azurerm_firewall" "example" {
  name                = "testfirewall"
  location            = "West Europe"

  ip_configuration {
    name      = "configuration"
    subnet_id = "subnetid"
  }

  resource_group_name = "rg"
  sku_name            = "AZFW_VNet"
  sku_tier            = "Standard"
}

# Explanation:
# The ip_configuration block appears, and after that we still have
# resource_group_name, sku_name, and so on as normal attributes.
# This is invalid because we have attributes following a nested block.
# Our rule should emit an issue for each attribute that’s out of order.
# But in the test we expect to see at least one issue complaining that
# "Argument 'resource_group_name' must not come after a nested block".

# The test references line 15 for resource_group_name => hcl.Pos{Line: 15, ...}
