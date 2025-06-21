resource "azurerm_storage_account" "this" {
  name                     = "storageaccount"
  resource_group_name      = "rg"
  location                 = "westeurope"
  account_tier             = "Premium"
  account_replication_type = "LRS"

  blob_properties {
    delete_retention_policy {
      days = 7
    }

    container_delete_retention_policy {
      days = 7
    }
    versioning_enabled = false
  }
}
