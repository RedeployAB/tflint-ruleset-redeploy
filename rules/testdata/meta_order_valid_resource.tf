resource "aws_instance" "example" {
  provider = aws
  count    = 1

  lifecycle {}

  depends_on = []
}
