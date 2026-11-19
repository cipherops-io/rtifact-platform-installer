data "terraform_remote_state" "core_network" {
  backend = var.upstream_state["core-network"].backend
  config  = var.upstream_state["core-network"].config
}

locals {
  core = data.terraform_remote_state.core_network.outputs
}

locals {
  data_services_enabled = (
    var.features.postgresql_flexible_server ||
    var.features.container_registry ||
    var.features.storage_account ||
    var.features.redis_cache ||
    var.features.azure_managed_redis ||
    var.features.azure_files_share
  )
}

module "security" {
  count = var.features.key_vault ? 1 : 0

  source              = "../../../../modules/azure/security"
  resource_group_name = local.core.resource_group_name
  location            = var.location
  name_prefix         = var.name_prefix
  tags                = local.core.tags
  sku                 = var.key_vault.sku
  tenant_id           = data.azurerm_client_config.current.tenant_id
  features = {
    disk_encryption_set = var.features.disk_encryption_set
  }
}

module "data" {
  count = local.data_services_enabled ? 1 : 0

  source              = "../../../../modules/azure/data"
  resource_group_name = local.core.resource_group_name
  location            = var.location
  name_prefix         = var.name_prefix
  tags                = local.core.tags
  features = {
    postgresql_flexible_server = var.features.postgresql_flexible_server
    container_registry         = var.features.container_registry
    storage_account            = var.features.storage_account
    redis_cache                = var.features.redis_cache
    azure_managed_redis        = var.features.azure_managed_redis
    private_dns_zones          = var.features.private_dns_zones
    azure_files_share          = var.features.azure_files_share
    disk_encryption_set        = false
  }
  managed_redis_sku_name           = var.azure_managed_redis.sku_name
  postgresql_version               = var.postgresql_flexible_server.version
  postgresql_sku_name              = var.postgresql_flexible_server.sku_name
  postgresql_storage_mb            = var.postgresql_flexible_server.storage_mb
  postgresql_storage_iops          = try(var.postgresql_flexible_server.iops, null)
  postgresql_ha_enabled            = try(var.postgresql_flexible_server.high_availability.enabled, false)
  postgresql_ha_mode               = try(var.postgresql_flexible_server.high_availability.mode, "ZoneRedundant")
  postgresql_read_replica_enabled  = try(var.postgresql_flexible_server.read_replica.enabled, false)
  postgresql_backup_retention_days = try(var.postgresql_flexible_server.backup_retention_days, 7)
  redis_sku_tier                   = try(var.redis_cache.sku_name, "Premium")
  redis_family                     = try(var.redis_cache.family, "P")
  redis_capacity                   = try(var.redis_cache.capacity, 1)
  redis_minimum_tls_version        = try(var.redis_cache.minimum_tls_version, "1.2")
  storage_corpus_gb                = var.storage_account.corpus_gb
  storage_backup_gb                = var.storage_account.backup_gb
  azure_files_share_name           = try(var.azure_files.share_name, "aios-shared")
  azure_files_quota_gb             = try(var.azure_files.quota_gb, 1024)
  azure_files_zone_redundant       = try(var.azure_files.zone_redundant, true)
  acr_sku                          = var.container_registry.sku
  private_endpoint_subnet_id       = local.core.private_endpoint_subnet_id
  vnet_id                          = local.core.vnet_id
  dns_resource_group_name          = local.core.resource_group_name
}

data "azurerm_client_config" "current" {}
