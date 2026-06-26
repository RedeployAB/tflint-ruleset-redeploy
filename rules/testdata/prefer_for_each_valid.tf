resource "aws_instance" "toggle" {
  count = var.create ? 1 : 0
}

resource "aws_instance" "literal_toggle" {
  count = 1
}

resource "aws_instance" "guarded_toggle" {
  count = length(var.subnets) > 0 ? 1 : 0
}

resource "aws_instance" "reference" {
  count = var.instance_count
}

resource "aws_instance" "for_each" {
  for_each = var.subnets
}
