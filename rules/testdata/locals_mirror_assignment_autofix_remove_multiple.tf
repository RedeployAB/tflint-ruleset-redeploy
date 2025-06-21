variable "env" {
  default = "dev"
}
variable "region" {
  default = "us-east-1"
}

locals {
  environment = var.env
  aws_region  = var.region
  computed    = lower(var.env)
}
