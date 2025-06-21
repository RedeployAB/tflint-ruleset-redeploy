variable "complex" {
  description = "Complex example"
  type        = list(string)
  default     = ["a", "b"]
  sensitive   = true
  nullable    = false
}
