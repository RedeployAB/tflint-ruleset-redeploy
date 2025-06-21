resource "test" "example" {
  name = "test"

  lifecycle {
    prevent_destroy = true
  }
}
