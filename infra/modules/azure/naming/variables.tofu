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