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

variable "features" {
  type = object({
    disk_encryption_set = optional(bool, false)
    ai_foundry          = optional(bool, false)
  })
  default = { disk_encryption_set = false, ai_foundry = false }
}

variable "foundry_key_vault" {
  description = "Foundry-dedicated Key Vault (access-policy model); set from customer key_vault.foundry / ai_foundry.key_vault YAML."
  type = object({
    sku                              = string
    soft_delete_retention_days       = number
    purge_protection_enabled         = bool
    grant_deployer_contributor       = bool
    deployer_key_permissions         = list(string)
    deployer_secret_permissions      = list(string)
    deployer_certificate_permissions = list(string)
  })
}