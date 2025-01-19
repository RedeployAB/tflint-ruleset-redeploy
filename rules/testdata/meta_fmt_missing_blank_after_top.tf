resource "aws_instance" "example" {
  provider = aws
  # next line isn't blank
  name = "test"
}
