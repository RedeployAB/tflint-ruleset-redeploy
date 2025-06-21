resource "random_uuid" "test" {
  for_each = local.test


  lifecycle {
    prevent_destroy = true
  }
}
