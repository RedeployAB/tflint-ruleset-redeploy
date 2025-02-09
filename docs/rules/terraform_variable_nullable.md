# terraform_variable_nullable

## What does this rule do?

This rule enforces proper handling of the `nullable` attribute in variable
blocks. It checks that:

- **If a variable has `default = null`, then the `nullable` attribute must not
  be declared.**
- **If the `nullable` attribute is declared, it must be set to `false` because
  `true` is redundant** (the default behavior is already nullable).
- **For boolean variables, setting `default = null` is not allowed.**

## Why is this important?

Correct usage of `nullable` ensures clarity about the expected input for a
variable and prevents conflicting or redundant declarations.

## How to fix issues

- **If your variable has `default = null`, remove any explicit `nullable`
  attribute.**

  **Incorrect:**

  ```hcl
  variable "example" {
    description = "An example variable"
    type        = string
    default     = null
    nullable    = true
  }
  ```

  **Correct:**

  ```hcl
  variable "example" {
    description = "An example variable"
    type        = string
    default     = null
  }
  ```

- **If you have declared `nullable = true`, remove it since the default is
  already `true`.**

  **Incorrect:**

  ```hcl
  variable "example" {
    description = "An example variable"
    type        = string
    nullable    = true
  }
  ```

  **Correct:**

  ```hcl
  variable "example" {
    description = "An example variable"
    type        = string
  }
  ```

- **For boolean variables, avoid setting `default = null`.**

  **Incorrect:**

  ```hcl
  variable "enable_feature" {
    description = "Enable feature flag"
    type        = bool
    default     = null
  }
  ```

  **Correct:**

  ```hcl
  variable "enable_feature" {
    description = "Enable feature flag"
    type        = bool
    default     = false  # or true, depending on the desired default
  }
  ```
