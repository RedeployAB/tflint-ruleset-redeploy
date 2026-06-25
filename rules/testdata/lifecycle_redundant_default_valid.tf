resource "aws_instance" "this" {
  lifecycle {
    prevent_destroy       = true
    create_before_destroy = true
  }
}

resource "aws_instance" "from_variable" {
  lifecycle {
    prevent_destroy = var.protect
  }
}

resource "aws_instance" "no_lifecycle" {
  ami = var.ami_id
}
