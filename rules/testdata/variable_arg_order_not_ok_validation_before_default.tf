variable "fail_validation_before_default" {
  description = "Another out-of-order"

  # 'validation' appears before 'default' => out-of-order
  validation {
    condition     = true
    error_message = "Should appear after default"
  }

  default = "some_value"
  # => Should fail the check with "Out-of-order argument 'validation'…"
}
