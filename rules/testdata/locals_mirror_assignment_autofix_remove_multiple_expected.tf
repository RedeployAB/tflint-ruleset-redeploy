variable "env" {
  default = "dev"
}
variable "region" {
  default = "us-east-1"
}

locals {
  computed = lower(var.env)
}
