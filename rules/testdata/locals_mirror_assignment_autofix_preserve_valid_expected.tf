variable "hello" {}

locals {
  hello = lower(var.hello)
}
