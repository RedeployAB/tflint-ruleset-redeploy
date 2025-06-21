variable "env" {
  default = "dev"
}

locals {
  env = var.env
}
