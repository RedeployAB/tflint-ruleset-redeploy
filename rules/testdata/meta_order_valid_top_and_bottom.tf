resource "aws_instance" "example" {
  provider = aws.west
  count    = 1

  ami           = "ami-123456"
  instance_type = "t2.micro"

  lifecycle {}

  depends_on = []
}
