terraform {
  required_providers {
    # AWS provider for cloud resources
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
    # Azure provider
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0"
    }
  }
}
