resource "random_id" "example" {
  // This is allowed now
  byte_length = 8
  keepers = {
    env = var.env_name
  }
}
