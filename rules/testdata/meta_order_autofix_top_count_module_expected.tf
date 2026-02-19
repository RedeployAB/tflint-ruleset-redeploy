module "example" {
  count = 2

  source  = "hashicorp/consul/aws"
  version = "0.1.0"
}
