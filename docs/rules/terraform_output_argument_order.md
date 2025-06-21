# terraform_output_argument_order

## What does this rule do?

This rule enforces a specific order for attributes and blocks within an output
block. The expected order is:

1. **description** (attribute)
2. **value** (attribute)
3. **ephemeral** (attribute)
4. **sensitive** (attribute)
5. **precondition** (block)
6. **depends_on** (attribute)

If any of these elements appears out of order, an error is emitted.

## Why is this important?

A consistent order for attributes and blocks in output blocks improves
readability and makes it easier for developers to quickly locate and understand
each output's purpose.

## How to fix issues

Rearrange the attributes and blocks in your output block so that they appear in
the order specified above.

### Example

**Before:**

```hcl
output "instance_id" {
  depends_on = [aws_instance.example]
  sensitive = true
  precondition {
    condition     = var.instance_count > 0
    error_message = "Must have at least one instance"
  }
  value = aws_instance.example.id
  description = "The instance ID"
}
```

**After:**

```hcl
output "instance_id" {
  description = "The instance ID"
  value       = aws_instance.example.id
  sensitive   = true

  precondition {
    condition     = var.instance_count > 0
    error_message = "Must have at least one instance"
  }

  depends_on = [aws_instance.example]
}
```
