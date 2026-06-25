resource "aws_instance" "this" {
  lifecycle {
    prevent_destroy       = false
    create_before_destroy = false
  }
}
