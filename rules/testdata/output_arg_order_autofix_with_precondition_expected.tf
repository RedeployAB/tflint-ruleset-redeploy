output "with_precondition" {
  description = "Output with validation"
  value       = var.test

  precondition {
    condition     = var.test != ""
    error_message = "Test must not be empty"
  }
}
