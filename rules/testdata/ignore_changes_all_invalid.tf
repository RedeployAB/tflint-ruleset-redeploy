resource "aws_instance" "this" {
  lifecycle {
    ignore_changes = all
  }
}
