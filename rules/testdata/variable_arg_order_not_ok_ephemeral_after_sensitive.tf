variable "fail_ephemeral_after_sensitive" {
  description = "Testing argument order"

  type      = string
  default   = "some_value"
  sensitive = true
  ephemeral = true
  # 'ephemeral' appears after 'sensitive', which is out-of-order.
  # The rule should raise an issue:
  # "Out-of-order argument 'ephemeral'. Expected sequence: description, type, default, ephemeral, sensitive, nullable, validation"
}
