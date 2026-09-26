output "multiple_preconditions" {
  description = "Output with several validations"
  value       = var.test

  precondition {
    condition     = var.test != ""
    error_message = "Test must not be empty"
  }

  precondition {
    condition     = length(var.test) < 64
    error_message = "Test must be shorter than 64 characters"
  }
}
