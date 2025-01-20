resource "azurerm_firewall_application_rule_collection" "example" {
  name                = "testcollection"
  azure_firewall_name = "myfw"
  resource_group_name = "rg"
  priority            = 100
  action              = "Allow"

  rule {
    name = "testrule"

    source_addresses = [
      "10.0.0.0/16",
    ]

    target_fqdns = [
      "*.google.com",
    ]

    protocol {
      port = "443"
      type = "Https"
    }
  }
}

# Explanation: normal attributes come first, then the 'rule' block,
# inside that we also have a 'protocol' block, but that's okay since
# it appears after the 'rule' block’s own normal attributes => valid
