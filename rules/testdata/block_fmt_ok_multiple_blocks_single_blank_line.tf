resource "azurerm_firewall_policy_rule_collection_group" "example" {
  application_rule_collection {
    name   = "app_rule_collection1"
    action = "Deny"
  }

  network_rule_collection {
    name   = "network_rule_collection1"
    action = "Deny"
  }

  nat_rule_collection {
    name   = "nat_rule_collection1"
    action = "Dnat"
  }
}
