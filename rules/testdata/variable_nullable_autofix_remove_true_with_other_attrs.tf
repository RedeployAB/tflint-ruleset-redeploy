variable "test" {
  description = "test variable"
  type        = string
  nullable    = true
  validation {
    condition     = length(var.test) > 0
    error_message = "Must not be empty"
  }
}
