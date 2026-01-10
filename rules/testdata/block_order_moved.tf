terraform {}

moved {
  from = aws_instance.old
  to   = aws_instance.new
}

resource "aws_instance" "new" {}
