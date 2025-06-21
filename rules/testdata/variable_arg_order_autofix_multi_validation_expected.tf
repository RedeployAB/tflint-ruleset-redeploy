variable "multi_validation" {
  type    = string
  default = "value"

  validation {
    condition     = true
    error_message = "First"
  }
  validation {
    condition     = true
    error_message = "Second"
  }
}
