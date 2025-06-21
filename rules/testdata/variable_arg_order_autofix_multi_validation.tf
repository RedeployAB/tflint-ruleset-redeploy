variable "multi_validation" {
  type = string
  validation {
    condition     = true
    error_message = "First"
  }
  default = "value"
  validation {
    condition     = true
    error_message = "Second"
  }
}
