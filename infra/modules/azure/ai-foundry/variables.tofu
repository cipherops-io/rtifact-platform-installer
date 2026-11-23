variable "resource_group_name" { type = string }
variable "location" { type = string }
variable "name_prefix" { type = string }
variable "storage_account_id" { type = string }
variable "key_vault_id" { type = string }
variable "container_registry_id" {
  type    = string
  default = null
}
variable "public_network_access" {
  type     = string
  default  = "Disabled"
  nullable = false

  validation {
    condition     = contains(["Enabled", "Disabled"], var.public_network_access)
    error_message = "ai_foundry.public_network_access must be Enabled or Disabled."
  }
}
variable "tags" {
  type    = map(string)
  default = {}
}