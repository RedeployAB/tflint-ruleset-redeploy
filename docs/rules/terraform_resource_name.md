# terraform_resource_name

## What does this rule do?

This rule checks that the name given to a resource does not redundantly include
its resource type. For example, in a resource of type `aws_instance`, the name
should not also include the word “instance.”

## Why is this important?

Avoiding redundant naming keeps resource identifiers concise and clear. It
reduces visual clutter and helps prevent confusion when managing or searching
for resources.

## How to fix issues

Rename the resource to remove the repeated type. For example, change:

```hcl
resource "aws_instance" "aws_instance" { ... }
```

to something like:

```hcl
resource "aws_instance" "example" { ... }
```
