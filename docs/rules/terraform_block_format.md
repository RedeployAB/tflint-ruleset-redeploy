# terraform_block_format

## What does this rule do?

This rule enforces consistent spacing between items (attributes and nested blocks) within a Terraform block. Specifically, it checks that:

- If a block contains no attributes, its first nested block appears immediately after the opening brace (with no blank lines).
- If a block has attributes, the first nested block must be preceded by exactly one blank line.
- Any subsequent nested blocks in the same block are also preceded by exactly one blank line.

If the spacing does not conform to these rules, an error is emitted.

## Why is this important?

Consistent spacing makes Terraform code more readable and maintainable. It visually distinguishes attributes from nested blocks and ensures that the structure is clear to anyone reviewing the code.

## How to fix issues

When an issue is reported:

- **No attributes present:** Remove any blank lines before the first nested block.
  
  ```hcl
  resource "aws_instance" "example" {
    tags = {
      Name = "example"
    }
  }
  ```

- **Attributes present:** Ensure there is exactly one blank line before the first nested block.

  ```hcl
  resource "aws_instance" "example" {
    ami           = "ami-123456"
    instance_type = "t2.micro"

    tags = {
      Name = "example"
    }
  }
  ```

- **Multiple nested blocks:** Ensure there is exactly one blank line separating each nested block.

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
