module "example" {
  source  = "hashicorp/consul/aws"
  version = "0.1.0"

  count = 2
}
