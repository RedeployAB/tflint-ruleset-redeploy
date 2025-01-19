resource "aws_nat_gateway" "this" {
  count = 2

  tags = {
    Name = "..."
  }

  allocation_id = "..."
}
