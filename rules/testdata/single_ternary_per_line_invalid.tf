locals {
  # Nested ternary on a single line: two ternary operations.
  nat_gateway_count = var.single_nat_gateway ? 1 : var.one_per_az ? length(var.azs) : local.max_subnet_length

  # Chained ternary on a single line: two ternary operations.
  release = var.ami_id != "" ? null : var.use_latest ? local.latest : var.release
}
