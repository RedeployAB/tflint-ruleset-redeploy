resource "aws_instance" "example" {
  provider = aws.west
  count    = 1
}
