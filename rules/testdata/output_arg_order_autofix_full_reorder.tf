output "full_reorder" {
  depends_on = [aws_instance.example]
  sensitive  = true
  precondition {
    condition     = var.test != ""
    error_message = "Test required"
  }
  value       = var.test
  description = "Full test"
  ephemeral   = true
}
