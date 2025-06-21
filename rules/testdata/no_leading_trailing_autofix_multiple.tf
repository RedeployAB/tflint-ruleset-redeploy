resource "test" "one" {

  name = "one"
}

module "two" {
  source = "./two"

}
