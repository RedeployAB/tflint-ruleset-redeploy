resource "aws_instance" "example" {
  ami           = "ami-123456"
  instance_type = "t2.micro"

  connection {
    type = "ssh"
    user = "ec2-user"
  }

  tags = {
    Name = "test"
  }
}
