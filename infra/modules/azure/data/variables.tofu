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

  validation {
    condition     = can(regex("^(Balanced_B|ComputeOptimized_X|FlashOptimized_A|MemoryOptimized_M)[0-9]+$", var.managed_redis_sku_name))
    error_message = "azure_managed_redis.sku_name must be a Managed Redis SKU such as Balanced_B3, ComputeOptimized_X5 or MemoryOptimized_M10."
  }
}

variable "postgresql_version" {
  type    = string
  default = "16"

  validation {
    condition     = contains(["11", "12", "13", "14", "15", "16", "17", "18"], var.postgresql_version)
    error_message = "postgresql_flexible_server.version must be one of 11-18."
  }
}

variable "postgresql_sku_name" {
  type    = string
  default = "GP_Standard_D8ds_v5"

  validation {
    condition     = can(regex("^(B|GP|MO)_Standard_[A-Za-z0-9_]+$", var.postgresql_sku_name))
    error_message = "postgresql_flexible_server.sku_name must follow tier_name, e.g. B_Standard_B1ms, GP_Standard_D4ds_v5 or MO_Standard_E4ds_v5."
  }
}

variable "postgresql_storage_mb" {
  type    = number
  default = 1048576

  validation {
    condition     = contains([32768, 65536, 131072, 262144, 524288, 1048576, 2097152, 4193280, 4194304, 8388608, 16777216, 33553408], var.postgresql_storage_mb)
    error_message = "postgresql_flexible_server.storage_mb must be a supported size (32768 ... 33553408)."
  }
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

  validation {
    condition     = contains(["SameZone", "ZoneRedundant"], var.postgresql_ha_mode)
    error_message = "postgresql_flexible_server.high_availability.mode must be SameZone or ZoneRedundant."
  }
}

variable "postgresql_read_replica_enabled" {
  type    = bool
  default = false
}

variable "postgresql_backup_retention_days" {
  type    = number
  default = 7

  validation {
    condition     = var.postgresql_backup_retention_days >= 7 && var.postgresql_backup_retention_days <= 35
    error_message = "postgresql_flexible_server.backup_retention_days must be between 7 and 35."
  }
}

variable "redis_sku_tier" {
  type    = string
  default = "Premium"

  validation {
    condition     = contains(["Basic", "Standard", "Premium"], var.redis_sku_tier)
    error_message = "redis_cache.sku_name must be Basic, Standard or Premium."
  }
}

variable "redis_family" {
  type    = string
  default = "P"

  validation {
    condition     = (var.redis_family == "P") == (var.redis_sku_tier == "Premium") && contains(["C", "P"], var.redis_family)
    error_message = "redis_cache.family must be P for Premium and C for Basic/Standard."
  }
}

variable "redis_capacity" {
  type    = number
  default = 1

  validation {
    condition     = var.redis_family == "P" ? contains([1, 2, 3, 4, 5], var.redis_capacity) : contains([0, 1, 2, 3, 4, 5, 6], var.redis_capacity)
    error_message = "redis_cache.capacity must be 1-5 for family P or 0-6 for family C."
  }
}

variable "redis_minimum_tls_version" {
  type    = string
  default = "1.2"

  validation {
    condition     = var.redis_minimum_tls_version == "1.2"
    error_message = "redis_cache.minimum_tls_version must be 1.2 (the only value Azure accepts)."
  }
}

variable "azure_files_share_name" {
  type    = string
  default = "aios-shared"

  validation {
    condition     = can(regex("^[a-z0-9](?:[a-z0-9]|-[a-z0-9]){2,62}$", var.azure_files_share_name))
    error_message = "azure_files.share_name must be 3-63 lowercase letters, digits or single hyphens."
  }
}

variable "azure_files_quota_gb" {
  type    = number
  default = 1024

  validation {
    condition     = var.azure_files_quota_gb >= 100 && var.azure_files_quota_gb <= 102400
    error_message = "azure_files.quota_gb must be between 100 and 102400 (premium file shares)."
  }
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

  validation {
    condition     = contains(["Basic", "Standard", "Premium"], var.acr_sku)
    error_message = "container_registry.sku must be Basic, Standard or Premium."
  }
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

  validation {
    condition     = contains(["Standard", "Premium"], var.foundry_storage.account_tier)
    error_message = "ai_foundry.storage.account_tier must be Standard or Premium."
  }
  validation {
    condition     = contains(["LRS", "GRS", "RAGRS", "ZRS", "GZRS", "RAGZRS"], var.foundry_storage.account_replication_type)
    error_message = "ai_foundry.storage.account_replication_type must be one of LRS, GRS, RAGRS, ZRS, GZRS or RAGZRS."
  }
  validation {
    condition     = contains(["Enabled", "Disabled", "SecuredByPerimeter"], var.foundry_storage.public_network_access)
    error_message = "ai_foundry.storage.public_network_access must be Enabled, Disabled or SecuredByPerimeter."
  }
}

variable "postgresql" {
  description = "PostgreSQL Flexible Server settings beyond SKU and storage (postgresql_flexible_server in customer YAML)."
  type = object({
    administrator_login           = optional(string, "aiosadmin")
    zone                          = optional(string, "1")
    read_replica_zone             = optional(string, "2")
    extensions                    = optional(list(string), ["CITEXT", "UUID-OSSP"])
    public_network_access_enabled = optional(bool, false)
    password_length               = optional(number, 24)
    password_special              = optional(bool, true)
  })
  default  = {}
  nullable = false

  validation {
    condition     = can(regex("^[a-z][a-z0-9_]{0,62}$", var.postgresql.administrator_login)) && !startswith(var.postgresql.administrator_login, "pg_") && !contains(["azure_superuser", "azure_pg_admin", "admin", "administrator", "root", "guest", "public"], var.postgresql.administrator_login)
    error_message = "postgresql_flexible_server.administrator_login must be 1-63 lowercase letters, digits or underscores, must not start with pg_ and must not be a reserved name (admin, root, azure_superuser, ...)."
  }
  validation {
    condition     = contains(["1", "2", "3", ""], var.postgresql.zone) && contains(["1", "2", "3", ""], var.postgresql.read_replica_zone)
    error_message = "postgresql_flexible_server.zone and read_replica_zone must be \"1\", \"2\", \"3\" or \"\" (no zone preference)."
  }
  validation {
    condition     = alltrue([for e in var.postgresql.extensions : can(regex("^[A-Z0-9_-]+$", e))])
    error_message = "postgresql_flexible_server.extensions must be upper-case extension names from the azure.extensions allow-list (for example CITEXT, UUID-OSSP, VECTOR)."
  }
  validation {
    condition     = var.postgresql.password_length >= 16 && var.postgresql.password_length <= 128
    error_message = "postgresql_flexible_server.password_length must be between 16 and 128."
  }
}

variable "redis_public_network_access_enabled" {
  description = "Allow public network access to Azure Cache for Redis (redis_cache.public_network_access_enabled)."
  type        = bool
  default     = false
  nullable    = false
}

variable "managed_redis" {
  description = "Azure Managed Redis settings (azure_managed_redis in customer YAML)."
  type = object({
    public_network_access = optional(string, "Disabled")
    client_protocol       = optional(string, "Encrypted")
    eviction_policy       = optional(string, "AllKeysLRU")
  })
  default  = {}
  nullable = false

  validation {
    condition     = contains(["Enabled", "Disabled"], var.managed_redis.public_network_access)
    error_message = "azure_managed_redis.public_network_access must be Enabled or Disabled."
  }
  validation {
    condition     = contains(["Encrypted", "Plaintext"], var.managed_redis.client_protocol)
    error_message = "azure_managed_redis.client_protocol must be Encrypted or Plaintext."
  }
  validation {
    condition     = contains(["AllKeysLFU", "AllKeysLRU", "AllKeysRandom", "VolatileLRU", "VolatileLFU", "VolatileTTL", "VolatileRandom", "NoEviction"], var.managed_redis.eviction_policy)
    error_message = "azure_managed_redis.eviction_policy must be one of AllKeysLFU, AllKeysLRU, AllKeysRandom, VolatileLRU, VolatileLFU, VolatileTTL, VolatileRandom or NoEviction."
  }
}

variable "storage" {
  description = "Corpus/backup storage account (storage_account in customer YAML)."
  type = object({
    account_tier             = optional(string, "Premium")
    account_replication_type = optional(string, "LRS")
    account_kind             = optional(string, "BlockBlobStorage")
    public_network_access    = optional(string, "Disabled")
    is_hns_enabled           = optional(bool, true)
    delete_retention_days    = optional(number, 30)
  })
  default  = {}
  nullable = false

  validation {
    condition     = contains(["Standard", "Premium"], var.storage.account_tier)
    error_message = "storage_account.account_tier must be Standard or Premium."
  }
  validation {
    condition     = contains(["LRS", "GRS", "RAGRS", "ZRS", "GZRS", "RAGZRS"], var.storage.account_replication_type)
    error_message = "storage_account.account_replication_type must be one of LRS, GRS, RAGRS, ZRS, GZRS or RAGZRS."
  }
  validation {
    condition     = contains(["BlobStorage", "BlockBlobStorage", "Storage", "StorageV2"], var.storage.account_kind)
    error_message = "storage_account.account_kind must be BlobStorage, BlockBlobStorage, Storage or StorageV2."
  }
  validation {
    condition     = var.storage.account_kind != "BlockBlobStorage" || var.storage.account_tier == "Premium"
    error_message = "storage_account.account_kind BlockBlobStorage requires account_tier Premium."
  }
  validation {
    condition     = var.storage.account_tier == "Standard" || contains(["LRS", "ZRS"], var.storage.account_replication_type)
    error_message = "Premium storage accounts only support LRS or ZRS replication."
  }
  validation {
    condition     = contains(["Enabled", "Disabled", "SecuredByPerimeter"], var.storage.public_network_access)
    error_message = "storage_account.public_network_access must be Enabled, Disabled or SecuredByPerimeter."
  }
  validation {
    condition     = var.storage.delete_retention_days >= 1 && var.storage.delete_retention_days <= 365
    error_message = "storage_account.delete_retention_days must be between 1 and 365."
  }
}

variable "azure_files" {
  description = "Azure Files share access (azure_files in customer YAML)."
  type = object({
    public_network_access = optional(string, "Disabled")
    enabled_protocol      = optional(string, "NFS")
  })
  default  = {}
  nullable = false

  validation {
    condition     = contains(["Enabled", "Disabled", "SecuredByPerimeter"], var.azure_files.public_network_access)
    error_message = "azure_files.public_network_access must be Enabled, Disabled or SecuredByPerimeter."
  }
  validation {
    condition     = contains(["NFS", "SMB"], var.azure_files.enabled_protocol)
    error_message = "azure_files.enabled_protocol must be NFS or SMB."
  }
}

variable "acr" {
  description = "Container registry access (container_registry in customer YAML). Only Premium gets a private endpoint; public access defaults to off for Premium and on otherwise."
  type = object({
    admin_enabled                 = optional(bool, false)
    public_network_access_enabled = optional(bool)
  })
  default  = {}
  nullable = false

  validation {
    condition     = var.acr_sku == "Premium" || var.acr.public_network_access_enabled != false
    error_message = "container_registry.public_network_access_enabled = false requires sku Premium (Basic and Standard registries are public-only)."
  }
}

variable "private_endpoint_connection" {
  description = "Approval settings for the data service private endpoints (network.private_endpoint)."
  type = object({
    is_manual_connection = optional(bool, false)
    request_message      = optional(string)
  })
  default  = {}
  nullable = false

  validation {
    condition     = var.private_endpoint_connection.request_message == null || (var.private_endpoint_connection.is_manual_connection && length(coalesce(var.private_endpoint_connection.request_message, "")) <= 140)
    error_message = "network.private_endpoint.request_message is only valid with is_manual_connection = true and must be at most 140 characters."
  }
}
