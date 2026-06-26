# terraform_ignore_changes_all

## What does this rule do?

This rule warns when a `lifecycle` block uses `ignore_changes = all` instead of
a list of specific attributes.

## Why is this important?

`ignore_changes = all` tells Terraform to ignore drift on every attribute of a
resource after creation. This silently masks configuration changes, including
ones you did not intend to ignore, and makes the resource's real state opaque
in plans. Listing the specific attributes to ignore keeps the intent explicit
and lets Terraform still reconcile everything else.

## How to fix issues

Replace `all` with the specific attributes that should be ignored.

**Incorrect:**

```hcl
resource "aws_instance" "this" {
  lifecycle {
    ignore_changes = all
  }
}
```

**Correct:**

```hcl
resource "aws_instance" "this" {
  lifecycle {
    ignore_changes = [tags, user_data]
  }
}
```
