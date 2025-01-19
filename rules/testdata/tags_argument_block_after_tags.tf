resource "aws_nat_gateway" "this" {
  tags = {
    Name = "..."
  }

  something_else {}
}
