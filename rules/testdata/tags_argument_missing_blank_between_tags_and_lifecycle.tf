resource "aws_nat_gateway" "this" {
  tags = {
    Name = "..."
  }
  lifecycle {}
}
