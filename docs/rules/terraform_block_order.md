# terraform_block_order

## What does this rule do?

This rule validates that top-level blocks in a Terraform file appear in the
following order:

1. **terraform**
2. **provider**
3. **data**
4. **resource**

If any of these blocks appear out of order, the rule emits an error.

## Blocks not subject to ordering

The following block types can appear anywhere in the file and are not subject
to ordering constraints:

- **variable** - Input variable declarations
- **output** - Output value declarations
- **locals** - Local value definitions
- **module** - Module calls
- **moved** - Resource move declarations (Terraform 1.1+)
- **import** - Import declarations (Terraform 1.5+)
- **removed** - Resource removal declarations (Terraform 1.7+)
- **check** - Check blocks for assertions (Terraform 1.5+)

These blocks are intentionally flexible in placement to allow for logical
grouping. For example, `moved` blocks are typically placed alongside the
resources they affect.

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

You can freely intersperse `moved`, `import`, `removed`, and `check` blocks
as needed:

```hcl
terraform {
  required_version = ">= 1.5.0"
}

provider "aws" {}

# Import block can appear here
import {
  to = aws_instance.imported
  id = "i-1234567890abcdef0"
}

data "aws_ami" "example" {}

# Moved block alongside related resource
moved {
  from = aws_instance.old_name
  to   = aws_instance.new_name
}

resource "aws_instance" "new_name" {
  ami           = data.aws_ami.example.id
  instance_type = "t2.micro"
}
```

Ensure your Terraform files adhere to this block order to resolve ordering
issues.
