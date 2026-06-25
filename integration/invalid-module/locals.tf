locals {
  a = 1

  nat_count = var.single ? 1 : var.per_az ? length(var.azs) : local.max
}
