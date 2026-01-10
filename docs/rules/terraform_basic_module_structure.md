# terraform_basic_module_structure

## What does this rule do?

This rule checks that a Terraform module contains the following required files:

- **main.tf**
- **variables.tf**
- **outputs.tf**
- **terraform.tf**

If any of these files is missing, the rule emits a warning.

Note that `locals.tf` is optional. If your module uses locals, they should be
placed in `locals.tf` (enforced by the `terraform_locals_file` rule), but the
file itself is not required if no locals are used.

## Why is this important?

Ensuring that these files exist provides a minimal structure for a Terraform
module, following the
[HashiCorp Standard Module Structure](https://developer.hashicorp.com/terraform/language/modules/develop/structure).

## How to fix issues

If an issue is reported for a missing file, add that file to your module. For
example, if the issue reports "Missing required file: `outputs.tf`", create an
`outputs.tf` file in your module root.
