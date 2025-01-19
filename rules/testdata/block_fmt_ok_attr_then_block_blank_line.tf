resource "azurerm_resource_provider_registration" "example" {
  name = "Microsoft.ContainerService"

  feature {
    name       = "AKS-DataPlaneAutoApprove"
    registered = true
  }
}
