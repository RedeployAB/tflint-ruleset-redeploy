resource "aws_instance" "example" {
  depends_on = []

  ami           = "ami-123456"
  instance_type = "t2.micro"
}
