variable "test" {
  description = "Test variable"
  type        = string
  default     = "value"

  validation {
    condition     = length(var.test) > 0
    error_message = "Must not be empty"
  }
}
