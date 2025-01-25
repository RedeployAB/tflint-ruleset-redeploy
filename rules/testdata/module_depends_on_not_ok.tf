module "my_module" {
  source     = "./some_module"
  depends_on = [aws_vpc.main]
}
