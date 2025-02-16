variable "location" {
  description = "The Azure Region."
}

variable "resource_group_name" {
  description = "The Resource Group name."
}

variable "tags" {
  description = "Mapping of tags to assign."
  type        = map(string)
  default     = {}
}
