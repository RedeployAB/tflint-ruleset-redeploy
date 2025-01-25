resource "random_id" "example" {

  byte_length = 8
  keepers = {
    env = var.env_name
  }
}
