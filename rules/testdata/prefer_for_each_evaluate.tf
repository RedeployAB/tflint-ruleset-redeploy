variable "instance_count" {
  description = "Number of instances."
  type        = number
  default     = 3
}

variable "enabled" {
  description = "Whether to create the instance."
  type        = bool
  default     = true
}

# Resolves to 3 via the default, so this should be flagged.
resource "aws_instance" "from_variable" {
  count = var.instance_count
}

# Arithmetic that resolves to 3, so this should be flagged.
resource "aws_instance" "from_arithmetic" {
  count = 1 + 2
}

# Resolves to a 0/1 toggle, so this should not be flagged.
resource "aws_instance" "toggle" {
  count = var.enabled ? 1 : 0
}
