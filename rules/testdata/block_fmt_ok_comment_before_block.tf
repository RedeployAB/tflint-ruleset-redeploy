resource "aws_instance" "example" {
  // Preceding comment before provisioner block
  provisioner "local-exec" {
    command = "echo 'Hello, World!'"
  }
}
