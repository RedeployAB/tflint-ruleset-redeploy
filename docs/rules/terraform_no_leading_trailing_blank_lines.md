# terraform_no_leading_trailing_blank_lines

## What does this rule do?

This rule ensures that there are no extra blank lines immediately after the opening brace `{` or immediately before the closing brace `}` of a block. This check applies to resource and module blocks to enforce a compact block structure.

## Why is this important?

Eliminating unnecessary blank lines helps maintain a clean, consistent style in your Terraform configurations. It makes the code more readable by reducing unnecessary whitespace and helps in identifying the contents of a block at a glance.

## How to fix issues

Remove any blank line immediately following the opening brace or preceding the closing brace of your blocks.

**Incorrect:**
```hcl
resource "aws_instance" "example" {

  ami           = "ami-123456"
  instance_type = "t2.micro"

}
```

**Correct:**
```hcl
resource "aws_instance" "example" {
  ami           = "ami-123456"
  instance_type = "t2.micro"
}
```

Ensure that the opening brace `{` is followed by content on the next line without blank lines, and the content is immediately followed by the closing brace `}` without preceding blank lines.
