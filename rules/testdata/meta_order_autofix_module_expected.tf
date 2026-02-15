module "example" {
  source  = "hashicorp/consul/aws"
  version = "0.1.0"

  depends_on = [aws_vpc.main]
}
