output "test" {
  value = "test"

  precondition {
    condition     = length(var.name) > 0
    error_message = "Name must not be empty"
  }
}
