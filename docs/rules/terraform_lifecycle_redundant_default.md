# terraform_lifecycle_redundant_default

## What does this rule do?

This rule checks the `lifecycle` block for meta-arguments that are explicitly
set to their default value of `false`:

- `prevent_destroy = false`
- `create_before_destroy = false`

Both default to `false`, so setting them explicitly is redundant.

## Why is this important?

Declaring an argument with its default value adds noise and can mislead readers
into thinking the behaviour was deliberately configured. Omitting the argument
keeps the `lifecycle` block focused on the settings that actually change
behaviour. This mirrors the ruleset's other redundant-default checks (such as
`terraform_variable_sensitive` and `terraform_output_ephemeral`).

## How to fix issues

Remove the redundant argument. Only declare `prevent_destroy` or
`create_before_destroy` when setting them to `true`.

**Incorrect:**

```hcl
resource "aws_instance" "this" {
  lifecycle {
    prevent_destroy       = false
    create_before_destroy = false
  }
}
```

**Correct:**

```hcl
resource "aws_instance" "this" {
  lifecycle {
    create_before_destroy = true
  }
}
```

Non-literal expressions (for example `prevent_destroy = var.protect`) are not
reported.
