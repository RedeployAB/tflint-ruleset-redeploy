resource "aws_nat_gateway" "this" {
  tags = {
    Name = "..."
  }
  depends_on = [aws_internet_gateway.this]
}
