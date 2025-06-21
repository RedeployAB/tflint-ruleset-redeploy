output "full_reorder" {
  description = "Full test"
  value       = var.test
  ephemeral   = true
  sensitive   = true

  precondition {
    condition     = var.test != ""
    error_message = "Test required"
  }

  depends_on = [aws_instance.example]
}
