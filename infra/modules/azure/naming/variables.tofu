variable "prefix" {
  type        = string
  description = "Short name prefix for resources."
}

variable "location" {
  type        = string
  description = "Azure region."
}

variable "tags" {
  type        = map(string)
  description = "Additional resource tags."
  default     = {}
}

variable "create_resource_group" {
  type        = bool
  description = "When false, adopt existing {prefix}-rg (e.g. after bootstrap)."
  default     = true
}