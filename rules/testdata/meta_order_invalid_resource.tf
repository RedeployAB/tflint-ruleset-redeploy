resource "aws_instance" "example" {
  depends_on = []
  count      = 1
}
