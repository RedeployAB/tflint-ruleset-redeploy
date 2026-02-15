resource "aws_instance" "example" {
  ami           = "ami-123456"
  instance_type = "t2.micro"

  lifecycle {
    ignore_changes = [tags]
  }

  depends_on = [aws_vpc.main]
}
