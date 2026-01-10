terraform {}

provider "aws" {}

moved {
  from = aws_instance.old
  to   = aws_instance.renamed
}

data "aws_ami" "example" {}

import {
  to = aws_instance.imported
  id = "i-1234567890abcdef0"
}

resource "aws_instance" "renamed" {}
resource "aws_instance" "imported" {}

check "health" {
  assert {
    condition     = true
    error_message = "Always passes"
  }
}
