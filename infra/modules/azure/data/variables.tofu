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

variable "features" {
  type = object({
    postgresql_flexible_server = bool
    container_registry         = bool
    storage_account            = bool
    redis_cache                = bool
    azure_managed_redis        = optional(bool, false)
    private_dns_zones          = bool
    azure_files_share          = bool
    disk_encryption_set        = optional(bool, false)
    ai_foundry                 = optional(bool, false)
  })
}

variable "managed_redis_sku_name" {
  type    = string
  default = "Balanced_B3"
}

variable "postgresql_version" {
  type    = string
  default = "16"
}

variable "postgresql_sku_name" {
  type    = string
  default = "GP_Standard_D8ds_v5"
}

variable "postgresql_storage_mb" {
  type    = number
  default = 1048576
}

variable "postgresql_storage_iops" {
  type    = number
  default = null
}

variable "postgresql_ha_enabled" {
  type    = bool
  default = false
}

variable "postgresql_ha_mode" {
  type    = string
  default = "ZoneRedundant"
}

variable "postgresql_read_replica_enabled" {
  type    = bool
  default = false
}

variable "postgresql_backup_retention_days" {
  type    = number
  default = 7
}

variable "redis_sku_tier" {
  type    = string
  default = "Premium"
}

variable "redis_family" {
  type    = string
  default = "P"
}

variable "redis_capacity" {
  type    = number
  default = 1
}

variable "redis_minimum_tls_version" {
  type    = string
  default = "1.2"
}

variable "azure_files_share_name" {
  type    = string
  default = "aios-shared"
}

variable "azure_files_quota_gb" {
  type    = number
  default = 1024
}

variable "azure_files_zone_redundant" {
  type    = bool
  default = true
}

variable "storage_corpus_gb" {
  type    = number
  default = 250
}

variable "storage_backup_gb" {
  type    = number
  default = 250
}

variable "acr_sku" {
  type    = string
  default = "Premium"
}

variable "private_endpoint_subnet_id" {
  type = string
}

variable "vnet_id" {
  type = string
}

variable "dns_resource_group_name" {
  type = string
}

variable "foundry_storage" {
  description = "Non-HNS storage for Azure AI Foundry; set from customer ai_foundry.storage YAML."
  type = object({
    account_tier             = string
    account_replication_type = string
    public_network_access    = string
  })
}