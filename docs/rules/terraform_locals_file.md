# terraform_locals_file

## What does this rule do?

This rule verifies that local variable definitions are placed in a file named
**locals.tf**. If local variables are defined in other files or the naming
convention is not followed, an error is emitted.

## Why is this important?

Keeping local variable definitions in a dedicated, properly named file (i.e.,
`locals.tf`) makes your Terraform code easier to navigate and maintain. It
ensures that local values are centralized and can be easily located when
reviewing or modifying the configuration.

## How to fix issues

If an issue is reported:

- **Move all local variable definitions into a file named `locals.tf`** in your
  module's root directory.
- **Ensure the `locals.tf` file contains only `locals` blocks** and related
  comments or documentation.

For example, your `locals.tf` file should look like:

```hcl
locals {
  common_tags = {
    Project = "Example"
    Owner   = "Team A"
  }

  instance_count = var.enable_feature ? 5 : 2
}
```

Remove `locals` blocks from other `.tf` files to comply with this rule.
