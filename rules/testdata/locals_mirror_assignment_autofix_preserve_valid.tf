variable "hello" {}

locals {
  hello      = lower(var.hello)
  direct_bad = var.hello
}
