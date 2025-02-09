# terraform_variable_file

## What does this rule do?

This rule ensures that variable blocks are defined only in files whose names
follow the pattern:

- `variables.tf` or
- `variables.<area>.tf` (for example, `variables.prod.tf`).

## Why is this important?

Keeping variable declarations in dedicated files improves the organization of
your module and makes it easier to locate and manage variable definitions.

## How to fix issues

If a variable block is found in an invalid file (for example, in `main.tf`),
move it into a file named `variables.tf` or `variables.<area>.tf` (using
snake_case for names).

**Example:**

**Incorrect (`main.tf`):**

```hcl
variable "example" {
  description = "An example variable"
  type        = string
}
```

**Correct (`variables.tf`):**

```hcl
variable "example" {
  description = "An example variable"
  type        = string
}
```
