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
  nullable    = false

  validation {
    condition     = contains(["standard", "premium"], var.sku)
    error_message = "key_vault.sku must be standard or premium."
  }
}

variable "key_vault" {
  description = "Workload Key Vault settings (key_vault in customer YAML)."
  type = object({
    soft_delete_retention_days = optional(number, 7)
    purge_protection_enabled   = optional(bool, true)
    rbac_authorization_enabled = optional(bool, true)
  })
  default  = {}
  nullable = false

  validation {
    condition     = var.key_vault.soft_delete_retention_days >= 7 && var.key_vault.soft_delete_retention_days <= 90
    error_message = "key_vault.soft_delete_retention_days must be between 7 and 90."
  }
  validation {
    condition     = var.key_vault.rbac_authorization_enabled || !var.features.disk_encryption_set
    error_message = "key_vault.rbac_authorization_enabled must be true when features.disk_encryption_set is on (the DES key uses RBAC role assignments)."
  }
  validation {
    condition     = var.key_vault.purge_protection_enabled || !var.features.disk_encryption_set
    error_message = "key_vault.purge_protection_enabled must be true when features.disk_encryption_set is on (Azure requires it for DES keys)."
  }
}

variable "des_key" {
  description = "Disk encryption set key (disk_encryption_set.key in customer YAML)."
  type = object({
    key_type                  = optional(string, "RSA")
    key_size                  = optional(number, 3072)
    auto_key_rotation_enabled = optional(bool, false)
  })
  default  = {}
  nullable = false

  validation {
    condition     = contains(["RSA", "RSA-HSM"], var.des_key.key_type)
    error_message = "disk_encryption_set.key.key_type must be RSA or RSA-HSM (disk encryption sets do not accept EC keys)."
  }
  validation {
    condition     = contains([2048, 3072, 4096], var.des_key.key_size)
    error_message = "disk_encryption_set.key.key_size must be 2048, 3072 or 4096."
  }
  validation {
    condition     = var.des_key.key_type != "RSA-HSM" || var.sku == "premium"
    error_message = "RSA-HSM keys require key_vault.sku premium."
  }
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

  validation {
    condition     = contains(["standard", "premium"], var.foundry_key_vault.sku)
    error_message = "key_vault.foundry.sku must be standard or premium."
  }
  validation {
    condition     = var.foundry_key_vault.soft_delete_retention_days >= 7 && var.foundry_key_vault.soft_delete_retention_days <= 90
    error_message = "key_vault.foundry.soft_delete_retention_days must be between 7 and 90."
  }
}