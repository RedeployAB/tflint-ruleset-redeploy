locals {
  # Three chained ternaries on one line.
  triple = var.a ? 1 : var.b ? 2 : var.c ? 3 : 4

  # Two independent ternaries in a collection on one line.
  pair = [var.a ? 1 : 2, var.b ? 3 : 4]
}
