resource "azurerm_resource_provider_registration" "example" {

  feature {
    name       = "AKS-DataPlaneAutoApprove"
    registered = true
  }
}
