# terraform_variable_argument_order

## What does this rule do?

This rule enforces that the attributes within a variable block appear in a specific order. The expected sequence is:
1. **description**
2. **type**
3. **default**
4. **ephemeral**
5. **sensitive**
6. **nullable**
7. **validation**

Any of these attributes may be omitted, but if present, they must follow the specified order.

## Why is this important?

A consistent attribute order improves readability and maintainability of your variable definitions. It makes it easier for developers to review and update your configuration.

## How to fix issues

Rearrange the attributes in your variable block so that they follow the order above. For example, if your block is out of order, move the attributes so that "description" comes first, then "type", followed by "default", and so on.

**Example:**

**Incorrect:**
```hcl
variable "example" {
  default     = "value"
  description = "An example variable"
  type        = string
}
```

**Correct:**
```hcl
variable "example" {
  description = "An example variable"
  type        = string
  default     = "value"
}
```
