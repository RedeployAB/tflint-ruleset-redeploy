variable "example_var_full" {
  description = "Full usage"
  type        = bool
  default     = true
  sensitive   = true
  nullable    = false

  # Single validation block
  validation {
    condition     = true
    error_message = "Oops"
  }
  # Everything is in the correct order: description, type, default, sensitive, nullable, validation
  # => Should pass
}
