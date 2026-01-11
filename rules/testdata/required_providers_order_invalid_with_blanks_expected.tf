terraform {
  required_providers {
    azapi = {
      source  = "Azure/azapi"
      version = ">= 2.0.0"
    }

    azurerm = {
      source  = "hashicorp/azurerm"
      version = ">= 4.0.0"
    }

    tls = {
      source  = "hashicorp/tls"
      version = ">= 4.0.0"
    }
  }
}
