# terraform_block_order

## What does this rule do?

This rule validates that top-level blocks in a Terraform file appear in the
following order:

1. **terraform**
2. **provider**
3. **data**
4. **resource**

If any of these blocks appear out of order, the rule emits an error.

## Why is this important?

Maintaining a consistent block order improves the readability and predictability
of Terraform configuration files. It helps users quickly locate and understand
the different sections of the configuration.

## How to fix issues

Reorder your top-level blocks so they follow the expected sequence:

- Place the `terraform` block first.
- Follow with `provider` blocks.
- Then come the `data` blocks.
- And finally, the `resource` blocks.

For example:

```hcl
terraform {
  required_version = ">= 1.0.0"
}

provider "aws" {
  region = "us-west-2"
}

data "aws_ami" "example" {
  most_recent = true
  owners      = ["amazon"]
}

resource "aws_instance" "example" {
  ami           = data.aws_ami.example.id
  instance_type = "t2.micro"
}
```

Ensure your Terraform files adhere to this block order to resolve ordering
issues.
