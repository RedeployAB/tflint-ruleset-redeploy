variable "password" {
  description = "A secret value."
  type        = string
}

output "endpoint" {
  description = "The endpoint."
  value       = "https://example.com"
}

resource "aws_instance" "this" {
  lifecycle {
    create_before_destroy = true
  }
}
