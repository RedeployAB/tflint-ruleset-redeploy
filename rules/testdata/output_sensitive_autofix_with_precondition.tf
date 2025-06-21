output "test" {
  value     = "test"
  sensitive = false

  precondition {
    condition     = length(var.name) > 0
    error_message = "Name must not be empty"
  }
}
