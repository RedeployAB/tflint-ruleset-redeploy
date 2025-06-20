# terraform_output_resource

## What does this rule do?

This rule detects output blocks that reference an entire resource or data source instead of a specific attribute. Outputs should reference attributes like `id` or `arn` rather than the resource or data block itself.

## Why is this important?

Referencing an entire resource in an output can break module consumers when the resource schema changes and may expose unnecessary information. Explicitly selecting attributes keeps outputs stable and clear.

## How to fix issues

Update the `value` of the output to reference a particular attribute. For example:

```hcl
output "instance_id" {
  value = aws_instance.example.id
}
```

Avoid declarations like `value = aws_instance.example`.
