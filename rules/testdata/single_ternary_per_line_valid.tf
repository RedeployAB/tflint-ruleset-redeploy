locals {
  # A single ternary per line is fine.
  enabled = var.create ? 1 : 0

  # A ternary spread across multiple lines is still a single ternary.
  name = (
    var.create
    ? var.name
    : "default"
  )
}

resource "aws_instance" "this" {
  count = var.create ? 1 : 0

  ami           = var.ami_id
  instance_type = var.instance_type
}
