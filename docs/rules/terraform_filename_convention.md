# terraform_filename_convention

## What does this rule do?

This rule enforces a strict naming convention for Terraform files. Filenames must follow one of these patterns:

- `<name>.tf`
- `<name>.<area>.tf`

All characters must be lowercase, using snake_case (alphanumeric characters and underscores). If a filename does not match this pattern, an error is issued.

## Why is this important?

A consistent naming convention helps keep your Terraform project organized, makes it easier to identify file purposes at a glance, and reduces the chance for errors due to unexpected file names.

## How to fix issues

Rename any Terraform file that does not follow the required pattern:

- Use only lowercase letters, numbers, and underscores.
- Ensure the filename ends with `.tf`.
- Follow the pattern `<name>.tf` or `<name>.<area>.tf`.

**Examples:**

- Rename `Main.Example.tf` to `main.example.tf`.
- Rename `variables.TF` to `variables.tf`.
- Rename `My-Variables.tf` to `my_variables.tf`.

Proper filenames:

- `main.tf`
- `variables.tf`
- `outputs.tf`
- `network_security.tf`
- `data_sources.tf`
