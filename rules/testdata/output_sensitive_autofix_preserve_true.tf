output "secret" {
  description = "A secret output"
  value       = var.secret_value
  sensitive   = true
}
