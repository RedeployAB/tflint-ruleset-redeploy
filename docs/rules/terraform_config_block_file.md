# terraform_config_block_file

## What does this rule do?

This rule checks that any Terraform configuration block (i.e., the `terraform`
block that contains backend and provider settings) is placed in a file named
**terraform.tf**. If a `terraform` block is found in any other file (for
example, `main.tf`), an error is emitted.

## Why is this important?

Placing the `terraform` configuration block in its own file helps keep your
module organized. It separates the configuration settings (such as required
versions and backend configuration) from resource definitions and other parts of
the module.

## How to fix issues

If the rule reports an issue:

- **Move the `terraform` block into a file named `terraform.tf`** in your
  module’s root directory.

For example, ensure your `terraform.tf` file contains:

```hcl
terraform {
  required_version = ">= 1.0.0"

  backend "s3" {
    bucket = "my-terraform-state"
    key    = "state.tfstate"
    region = "us-west-2"
  }
}
```

Remove the `terraform` block from other files (e.g., `main.tf`) to adhere to
this convention.
