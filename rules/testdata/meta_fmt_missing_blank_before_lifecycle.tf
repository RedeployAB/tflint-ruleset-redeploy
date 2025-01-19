resource "aws_instance" "example" {
  tags = {
    Something = "xyz"
  }
  lifecycle {}
}
