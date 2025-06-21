output "with_precondition" {
  value = var.test
  precondition {
    condition     = var.test != ""
    error_message = "Test must not be empty"
  }
  description = "Output with validation"
}
