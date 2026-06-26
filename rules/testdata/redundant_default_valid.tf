variable "password" {
  description = "A secret value."
  type        = string
  sensitive   = true
}

output "endpoint" {
  description = "The endpoint."
  value       = "https://example.com"
}

resource "aws_instance" "this" {
  lifecycle {
    prevent_destroy       = true
    create_before_destroy = var.replace
  }
}
