resource "aws_instance" "example" {
  depends_on = [aws_vpc.main]

  ami           = "ami-123456"
  instance_type = "t2.micro"
}
