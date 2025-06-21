variable "complex" {
  nullable    = false
  description = "Complex example"
  sensitive   = true
  type        = list(string)
  default     = ["a", "b"]
}
