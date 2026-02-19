resource "aws_instance" "example" {
  lifecycle {}

  ami           = "ami-123456"
  instance_type = "t2.micro"

  for_each = var.instances
}
