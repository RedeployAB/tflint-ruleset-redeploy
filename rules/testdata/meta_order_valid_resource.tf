resource "aws_instance" "example" {
  count     = 1
  provider  = aws

  lifecycle {}

  depends_on = []
}
