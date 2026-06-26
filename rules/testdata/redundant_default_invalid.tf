variable "password" {
  description = "A secret value."
  type        = string
  sensitive   = false
}

output "endpoint" {
  description = "The endpoint."
  value       = "https://example.com"
  ephemeral   = false
}

resource "aws_instance" "this" {
  lifecycle {
    prevent_destroy       = false
    create_before_destroy = false
  }
}
