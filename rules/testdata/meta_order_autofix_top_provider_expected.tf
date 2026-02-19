resource "aws_instance" "example" {
  provider = aws.west

  ami           = "ami-123456"
  instance_type = "t2.micro"
}
