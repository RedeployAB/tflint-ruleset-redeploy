# terraform_block_format

## What does this rule do?

This rule enforces consistent spacing between items (attributes and nested
blocks) within Terraform blocks. It applies to all top-level Terraform blocks
(resource, data, terraform, provider, variable, output) and ALL their nested
blocks, including provider-specific blocks like `metric_query`, `ingress`,
`egress`, etc.

Specifically, it checks that:

- If a block's first item is a nested block (no attributes before it), the
  nested block appears immediately after the opening brace (with no empty lines).
  Comments are allowed and don't affect this requirement.
- If a block has attributes before its first nested block, the first nested
  block must be preceded by exactly one empty line.
- Any subsequent nested blocks in the same block are also preceded by exactly
  one empty line.

Note: Comment lines are not counted as empty lines and do not affect the
spacing requirements.

If the spacing does not conform to these rules, an error is emitted.

## Why is this important?

Consistent spacing makes Terraform code more readable and maintainable. It
visually distinguishes attributes from nested blocks and ensures that the
structure is clear to anyone reviewing the code.

## How to fix issues

When an issue is reported:

- **No attributes present:** Remove any empty lines before the first nested
  block. Comments are allowed.

  ```hcl
  resource "aws_instance" "example" {
    # Comments are allowed here
    tags = {
      Name = "example"
    }
  }

  lifecycle {
    # This comment is fine
    create_before_destroy = true
  }
  ```

- **Attributes present:** Ensure there is exactly one empty line before the
  first nested block.

  ```hcl
  resource "aws_instance" "example" {
    ami           = "ami-123456"
    instance_type = "t2.micro"

    tags = {
      Name = "example"
    }
  }
  ```

- **Nested block before attributes:** The nested block should appear
  immediately after the opening brace, even if attributes follow.

  ```hcl
  resource "azurerm_key_vault_key" "example" {
    name = "example-key"

    rotation_policy {
      automatic {
        time_before_expiry = "P30D"
      }

      expire_after         = "P90D"
      notify_before_expiry = "P29D"
    }
  }
  ```

- **Multiple nested blocks:** Ensure there is exactly one empty line separating
  each nested block.

  ```hcl
  resource "aws_security_group" "example" {
    name = "example-sg"

    ingress {
      from_port   = 80
      to_port     = 80
      protocol    = "tcp"
      cidr_blocks = ["0.0.0.0/0"]
    }

    egress {
      from_port   = 0
      to_port     = 0
      protocol    = "-1"
      cidr_blocks = ["0.0.0.0/0"]
    }
  }
  ```

Adjust your code to follow these spacing conventions to resolve the errors.
