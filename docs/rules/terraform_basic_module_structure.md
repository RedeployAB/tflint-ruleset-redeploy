# terraform_basic_module_structure

## What does this rule do?

This rule checks that a Terraform module contains the following required files:

- **main.tf**
- **variables.tf**
- **locals.tf**
- **outputs.tf**
- **terraform.tf**

If any of these files is missing, the rule emits a warning.

## Why is this important?

Ensuring that these files exist provides a minimal structure for a Terraform
module.

## How to fix issues

If an issue is reported for a missing file, add that file to your module. For
example, if the issue reports “Missing required file: `locals.tf`”, create a
`locals.tf` file in your module root.
