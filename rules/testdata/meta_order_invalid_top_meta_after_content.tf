resource "azurerm_role_assignment" "blob_contributor" {
  scope                = each.value.id
  role_definition_name = "Storage Blob Data Contributor"
  principal_id         = azurerm_databricks_access_connector.this.identity[0].principal_id

  for_each = var.storage_accounts
}
