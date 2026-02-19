resource "aws_instance" "example" {
  # Loop over instances
  for_each = var.instances

  ami           = "ami-123456"
  instance_type = "t2.micro"
}
