# terraform_variable_sensitive

## What does this rule do?

This rule validates the usage of the `sensitive` attribute in variable blocks. It ensures that if the attribute is present, it is not explicitly set to `false` because the default behavior is non-sensitive. In other words, if a variable is not sensitive, you should simply omit the attribute.

## Why is this important?

Explicitly setting `sensitive = false` is redundant and can cause confusion. The rule helps maintain clean and clear variable definitions.

## How to fix issues

Remove the `sensitive = false` assignment from your variable block. Only include the `sensitive` attribute when you intend to mark the variable as sensitive (by setting it to `true`).

**Example:**

**Incorrect:**
```hcl
variable "password" {
  description = "The database password"
  type        = string
  sensitive   = false
}
```

**Correct:**
```hcl
variable "password" {
  description = "The database password"
  type        = string
}
```

**If the variable is sensitive:**

```hcl
variable "password" {
  description = "The database password"
  type        = string
  sensitive   = true
}
```
