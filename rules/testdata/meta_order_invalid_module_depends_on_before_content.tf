module "example" {
  depends_on = []

  source  = "hashicorp/consul/aws"
  version = "0.1.0"
}
