variable "ephemeral_test" {
  description = "Test"
  type        = string
  sensitive   = true
  ephemeral   = true
  default     = "value"
}
