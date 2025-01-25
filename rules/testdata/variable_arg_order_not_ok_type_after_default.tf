variable "fail_type_after_default" {
  description = "Out-of-order example"
  default     = "some_value"
  type        = string
  # 'type' appears after 'default', which is out-of-order.
  # The rule should raise an issue:
  # "Out-of-order argument 'type'. Expected sequence: description, type, default, sensitive, nullable, validation"
  # => Should fail
}
