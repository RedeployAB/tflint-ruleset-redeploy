terraform {
  required_providers {
    # Random provider
    random = {
      source  = "hashicorp/random"
      version = "~> 3.0"
    }
    # AWS provider
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }
}
