resource "aws_instance" "example" {
  ami           = "ami-123456"
  instance_type = "t2.micro"

  # Loop over instances
  for_each = var.instances
}
