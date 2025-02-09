# terraform_meta_argument_format

## What does this rule do?

This rule enforces proper formatting of meta-arguments in resource and module
blocks. It ensures that:

- There is a empty line **after** the top meta-arguments (`count`, `for_each`,
  or `provider`) if there is additional content following.
- There is a empty line **before** bottom meta-arguments (`depends_on` or
  `lifecycle`) when they are present.

If the spacing does not meet these expectations, an error is emitted.

## Why is this important?

Consistent formatting of meta-arguments improves the readability of your
configuration by clearly separating meta-arguments from the rest of the block’s
content. This helps in quickly identifying the meta-arguments and understanding
the structure of your resources and modules.

## How to fix issues

Adjust your configuration to insert or remove empty lines to meet the expected
spacing:

- **Insert a empty line after top meta-arguments if missing.**

  **Incorrect:**

  ```hcl
  resource "aws_instance" "example" {
    count = var.instance_count
    ami   = "ami-123456"
    instance_type = "t2.micro"
  }
  ```

  **Correct:**

  ```hcl
  resource "aws_instance" "example" {
    count = var.instance_count

    ami           = "ami-123456"
    instance_type = "t2.micro"
  }
  ```

- **Insert a empty line before bottom meta-arguments if missing.**

  **Incorrect:**

  ```hcl
  resource "aws_instance" "example" {
    ami           = "ami-123456"
    instance_type = "t2.micro"
    depends_on    = [aws_vpc.example]
  }
  ```

  **Correct:**

  ```hcl
  resource "aws_instance" "example" {
    ami           = "ami-123456"
    instance_type = "t2.micro"

    depends_on = [aws_vpc.example]
  }
  ```

Ensure your resource and module blocks follow these formatting guidelines to
resolve issues reported by this rule.
