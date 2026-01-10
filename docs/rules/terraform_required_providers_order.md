# terraform_required_providers_order

## What does this rule do?

This rule checks that providers in the `required_providers` block are listed in
alphabetical order (case-insensitive).

## Why is this important?

Alphabetically ordering providers in the `required_providers` block improves
readability and makes it easier to find specific provider configurations. This
follows the Azure Verified Modules (AVM) specification TFNFR26.

## How to fix issues

Reorder the provider declarations in your `required_providers` block to be in
alphabetical order.

**Example:**

**Incorrect:**

```hcl
terraform {
  required_providers {
    random = {
      source  = "hashicorp/random"
      version = "~> 3.0"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0"
    }
  }
}
```

**Correct:**

```hcl
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.0"
    }
  }
}
```

This rule supports auto-fix. Running TFLint with the `--fix` flag will
automatically reorder the providers.
