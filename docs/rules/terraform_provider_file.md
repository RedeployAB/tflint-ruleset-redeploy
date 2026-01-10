# terraform_provider_file

## What does this rule do?

This rule checks that `provider` blocks are placed in files named `providers.tf`
or `providers.<area>.tf` (e.g., `providers.azure.tf`).

## Why is this important?

Placing provider configurations in dedicated files improves code organization
and makes it easier to locate and manage provider settings. This follows the
HashiCorp Terraform style guide recommendation for file organization.

Note: According to Azure Verified Modules (AVM) TFNFR27, reusable modules
should not contain provider blocks at all. Provider configurations should only
exist in root modules.

## How to fix issues

Move your `provider` blocks to a file named `providers.tf` or
`providers.<area>.tf`.

**Example:**

**Incorrect (provider in main.tf):**

```hcl
# main.tf
provider "aws" {
  region = "us-east-1"
}

resource "aws_instance" "example" {
  ami           = "ami-12345678"
  instance_type = "t2.micro"
}
```

**Correct (provider in providers.tf):**

```hcl
# providers.tf
provider "aws" {
  region = "us-east-1"
}
```

```hcl
# main.tf
resource "aws_instance" "example" {
  ami           = "ami-12345678"
  instance_type = "t2.micro"
}
```

You can also use area-specific provider files:

```hcl
# providers.azure.tf
provider "azurerm" {
  features {}
}
```
