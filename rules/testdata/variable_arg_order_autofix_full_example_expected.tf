variable "full_example" {
  description = "Full example"
  type        = string
  default     = "test"
  ephemeral   = true
  sensitive   = true
  nullable    = false

  validation {
    condition     = var.full_example != ""
    error_message = "Cannot be empty"
  }
}
