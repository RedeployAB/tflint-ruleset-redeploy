resource "aws_nat_gateway" "this" {
  allocation_id = "..."
  subnet_id     = "..."

  tags = {
    Name = "..."
  }
}
