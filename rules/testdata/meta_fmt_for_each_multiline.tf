resource "azurerm_subnet" "replica_set" {
  for_each = {
    for replica_set in local.replica_sets :
    replica_set.subnet_name => replica_set
  }

  name                 = each.value.subnet_name
  resource_group_name  = each.value.virtual_network_resource_group_name
  virtual_network_name = each.value.virtual_network_name
  address_prefixes     = [each.value.subnet_cidr]
}
