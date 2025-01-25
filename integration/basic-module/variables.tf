variable "location" {
  description = "The Azure Region in which all resources in this example should be created."
}

variable "resource_group_name" {
  description = "The name of the Resource Group in which all resources in this example should be created."
}

variable "tags" {
  description = "A mapping of tags to assign to the resources in this example."
  type        = map(string)
  default     = {}
}
