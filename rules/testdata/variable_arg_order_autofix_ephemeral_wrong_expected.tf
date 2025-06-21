variable "ephemeral_test" {
  description = "Test"
  type        = string
  default     = "value"
  ephemeral   = true
  sensitive   = true
}
