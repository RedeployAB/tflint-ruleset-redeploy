variable "example_var_minimal" {
  description = "A minimal variable."
  type        = string
  # No default, no sensitive, no nullable, no validation
  # This should pass the order check.
}
