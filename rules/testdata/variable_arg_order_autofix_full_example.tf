variable "full_example" {
  nullable = false
  validation {
    condition     = var.full_example != ""
    error_message = "Cannot be empty"
  }
  type        = string
  sensitive   = true
  description = "Full example"
  default     = "test"
  ephemeral   = true
}
