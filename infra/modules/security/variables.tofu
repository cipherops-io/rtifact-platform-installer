variable "resource_group_name" {
  type = string
}

variable "location" {
  type = string
}

variable "name_prefix" {
  type = string
}

variable "tags" {
  type    = map(string)
  default = {}
}

variable "sku" {
  type        = string
  description = "Key Vault SKU (standard or premium)."
  default     = "premium"
}

variable "tenant_id" {
  type = string
}