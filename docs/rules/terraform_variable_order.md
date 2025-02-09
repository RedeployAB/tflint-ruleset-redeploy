# terraform_variable_order

## What does this rule do?

This rule checks that variable declarations are ordered as follows:

1. **All required variables** (those without a default value) come first in
   **alphabetical order**.
2. Followed by **all optional variables** (those with a default value) in
   **alphabetical order**.

## Why is this important?

Ordering variables in a predictable way makes your module easier to read and
maintain. It helps users quickly identify required inputs versus optional ones.

## How to fix issues

Rearrange your variable blocks so that:

- **All required variables appear at the top and are sorted alphabetically.**
- **All optional variables appear after required ones and are also sorted
  alphabetically.**

**Example:**

**Incorrect:**

```hcl
variable "zzz_optional" {
  description = "An optional variable"
  type        = string
  default     = "value"
}

variable "aaa_required" {
  description = "A required variable"
  type        = string
}

variable "bbb_optional" {
  description = "Another optional variable"
  type        = string
  default     = "value"
}
```

**Correct:**

```hcl
# Required variables (alphabetical order)
variable "aaa_required" {
  description = "A required variable"
  type        = string
}

# Optional variables (alphabetical order)
variable "bbb_optional" {
  description = "Another optional variable"
  type        = string
  default     = "value"
}

variable "zzz_optional" {
  description = "An optional variable"
  type        = string
  default     = "value"
}
```
