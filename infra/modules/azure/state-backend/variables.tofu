variable "name_prefix" { type = string }
variable "location" { type = string }
variable "subscription_id" { type = string }
variable "resource_group_name" { type = string }

variable "create_resource_group" {
  type        = bool
  description = "When false, use an existing resource group (e.g. created by core-network or a prior bootstrap)."
  default     = true
}

variable "create_storage_account" {
  type        = bool
  description = "When false, use an existing storage account (pre-checked by installer)."
  default     = true
}

variable "storage_account" {
  description = "Remote state storage account settings (state.storage_account in customer YAML)."
  type = object({
    account_tier             = optional(string, "Standard")
    account_replication_type = optional(string, "LRS")
  })
  default  = {}
  nullable = false

  validation {
    condition     = contains(["Standard", "Premium"], var.storage_account.account_tier)
    error_message = "state.storage_account.account_tier must be Standard or Premium."
  }
  validation {
    condition     = contains(["LRS", "GRS", "RAGRS", "ZRS", "GZRS", "RAGZRS"], var.storage_account.account_replication_type)
    error_message = "state.storage_account.account_replication_type must be one of LRS, GRS, RAGRS, ZRS, GZRS or RAGZRS."
  }
  validation {
    condition     = var.storage_account.account_tier == "Standard" || contains(["LRS", "ZRS"], var.storage_account.account_replication_type)
    error_message = "Premium storage accounts only support LRS or ZRS replication."
  }
}

variable "create_container" {
  type        = bool
  description = "When false, use an existing storage container (pre-checked by installer)."
  default     = true
}