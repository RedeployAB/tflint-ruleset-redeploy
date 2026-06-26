# terraform_redundant_default

## What does this rule do?

This rule flags meta-arguments that are explicitly set to their default value of
`false`, which is redundant. It covers:

| Argument                | Context              |
| ----------------------- | -------------------- |
| `sensitive`             | `variable`, `output` |
| `ephemeral`             | `variable`, `output` |
| `prevent_destroy`       | `lifecycle`          |
| `create_before_destroy` | `lifecycle`          |

Only a literal `false` is reported. Non-literal expressions (for example
`prevent_destroy = var.protect`) are left alone.

## Why is this important?

Declaring an argument with its default value adds noise and can mislead readers
into thinking the behaviour was deliberately configured. Omitting the argument
keeps the configuration focused on the settings that actually change behaviour.

## How to fix issues

Remove the redundant argument. Only declare these arguments when setting them to
a non-default value. This rule supports autofix, which removes the redundant
line.

**Incorrect:**

```hcl
variable "password" {
  description = "A secret value."
  type        = string
  sensitive   = false
}

resource "aws_instance" "this" {
  lifecycle {
    prevent_destroy       = false
    create_before_destroy = false
  }
}
```

**Correct:**

```hcl
variable "password" {
  description = "A secret value."
  type        = string
}

resource "aws_instance" "this" {
  ami = var.ami_id
}
```

## Configuration

Each check can be disabled individually. All checks are enabled by default.

| Name                    | Default | Description                                                |
| ----------------------- | ------- | ---------------------------------------------------------- |
| `sensitive`             | `true`  | Check `sensitive = false` on variables and outputs.        |
| `ephemeral`             | `true`  | Check `ephemeral = false` on variables and outputs.        |
| `prevent_destroy`       | `true`  | Check `prevent_destroy = false` in lifecycle blocks.       |
| `create_before_destroy` | `true`  | Check `create_before_destroy = false` in lifecycle blocks. |

```hcl
rule "terraform_redundant_default" {
  enabled = true

  # Allow an explicit `create_before_destroy = false` without reporting it.
  create_before_destroy = false
}
```
