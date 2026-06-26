# terraform_provider_alias_order

## What does this rule do?

This rule enforces the ordering conventions for provider aliasing:

- In an aliased provider block, the `alias` argument must be the **first**
  argument.
- When multiple instances of the same provider are declared in a file, the
  default (un-aliased) provider block must be declared **before** any aliased
  instance.

## Why is this important?

Putting `alias` first makes it immediately clear that a provider block is a
non-default configuration. Declaring the default provider before its aliases
keeps the primary configuration easy to find. Both conventions come from the
[HashiCorp Terraform style guide](https://developer.hashicorp.com/terraform/language/style#provider-aliasing).

## How to fix issues

Move the `alias` argument to the top of the provider block, and declare the
default provider before any aliased instances.

**Incorrect:**

```hcl
provider "aws" {
  alias  = "us_east_1"
  region = "us-east-1"
}

provider "aws" {
  region = "eu-north-1"
}
```

**Correct:**

```hcl
provider "aws" {
  region = "eu-north-1"
}

provider "aws" {
  alias  = "us_east_1"
  region = "us-east-1"
}
```
