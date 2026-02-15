module "example" {
  depends_on = [aws_vpc.main]

  source  = "hashicorp/consul/aws"
  version = "0.1.0"
}
