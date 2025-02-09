# terraform_locals_mirror_assignment

## What does this rule do?

This rule checks for mirror assignments in `locals` blocks. It scans for local variable assignments that reference a variable directly (for example, `locals { foo = var.foo }`) without any transformation. Such mirror assignments are redundant and are discouraged.

## Why is this important?

Direct mirror assignments duplicate values unnecessarily and add extra maintenance overhead. Instead of mirroring a variable, you should reference the variable directly wherever needed. This promotes cleaner and more maintainable code.

## How to fix issues

If the rule reports an issue (e.g., "Local 'foo' is assigned directly from variable 'foo'"), remove the local assignment and update your code to reference `var.foo` directly.

**Example:**

**Incorrect:**
```hcl
locals {
  foo = var.foo
}
```

**Correct:**
```hcl
# Remove the local assignment and use var.foo directly in your code
resource "aws_instance" "example" {
  instance_type = var.instance_type
  # Use var.foo instead of local.foo
  ami           = lookup(var.ami_map, var.foo)
}
```
