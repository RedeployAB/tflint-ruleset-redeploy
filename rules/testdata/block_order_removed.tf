terraform {}

removed {
  from = aws_instance.old

  lifecycle {
    destroy = false
  }
}

resource "aws_instance" "new" {}
