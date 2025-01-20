provider "azurerm" {
  skip_provider_registration = true

  features {}
}

# Explanation: the non-block argument 'skip_provider_registration' comes first,
# then the 'features' block => valid
