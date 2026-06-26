resource "aws_instance" "literal" {
  count = 3
}

resource "aws_subnet" "from_length" {
  count = length(var.subnets)
}

resource "aws_subnet" "conditional_length" {
  count = var.create ? length(var.subnets) : 0
}
