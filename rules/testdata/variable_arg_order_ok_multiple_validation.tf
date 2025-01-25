variable "ok_multiple_validation" {
  description = "Testing multiple validations"
  type        = map(string)
  # No default, sensitive, or nullable => that's fine

  validation {
    condition     = length(var.ok_multiple_validation) > 0
    error_message = "Must not be empty"
  }

  # Multiple validation blocks, all after other attributes
  validation {
    condition     = var.ok_multiple_validation["some_key"] == "some_val"
    error_message = "some_key must be some_val"
  }
  # => Should pass
}
