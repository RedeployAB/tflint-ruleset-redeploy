terraform {}

resource "aws_instance" "example" {}

import {
  to = aws_instance.example
  id = "i-1234567890abcdef0"
}
