# terraform_variable_ephemeral

## What does this rule do?

This rule checks variable blocks for the usage of the `ephemeral` attribute. It
ensures that if the attribute is present, it is not explicitly set to `false`
(since `false` is the default behavior).

## Why is this important?

Setting `ephemeral = false` is redundant because the default behavior is already
non-ephemeral. Explicitly declaring it as `false` may cause confusion and
clutters the configuration.

## How to fix issues

Remove the `ephemeral = false` line from your variable block. Only include the
attribute when you want to set it to `true`.

**Example:**

**Incorrect:**

```hcl
variable "example" {
  description = "An example variable"
  type        = string
  ephemeral   = false
}
```

**Correct:**

```hcl
variable "example" {
  description = "An example variable"
  type        = string
}
```

**If you need the variable to be ephemeral:**

```hcl
variable "example" {
  description = "An example variable"
  type        = string
  ephemeral   = true
}
```
