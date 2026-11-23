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
    ai_foundry          = var.features.ai_foundry
  }
  foundry_key_vault = local.foundry_key_vault_settings
  key_vault = {
    soft_delete_retention_days = var.key_vault.soft_delete_retention_days
    purge_protection_enabled   = var.key_vault.purge_protection_enabled
    rbac_authorization_enabled = var.key_vault.rbac_authorization_enabled
  }
  des_key = {
    key_type                  = try(var.disk_encryption_set.key.key_type, null)
    key_size                  = try(var.disk_encryption_set.key.key_size, null)
    auto_key_rotation_enabled = var.disk_encryption_set.rotation_to_latest_version
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
    ai_foundry                 = var.features.ai_foundry
  }
  managed_redis_sku_name           = var.azure_managed_redis.sku_name
  postgresql_version               = var.postgresql_flexible_server.version
  postgresql_sku_name              = var.postgresql_flexible_server.sku_name
  postgresql_storage_mb            = var.postgresql_flexible_server.storage_mb
  postgresql_storage_iops          = try(var.postgresql_flexible_server.iops, null)
  postgresql_ha_enabled            = var.postgresql_flexible_server.high_availability.enabled
  postgresql_ha_mode               = var.postgresql_flexible_server.high_availability.mode
  postgresql_read_replica_enabled  = var.postgresql_flexible_server.read_replica.enabled
  postgresql_backup_retention_days = var.postgresql_flexible_server.backup_retention_days
  redis_sku_tier                   = var.redis_cache.sku_name
  redis_family                     = var.redis_cache.family
  redis_capacity                   = var.redis_cache.capacity
  redis_minimum_tls_version        = var.redis_cache.minimum_tls_version
  storage_corpus_gb                = var.storage_account.corpus_gb
  storage_backup_gb                = var.storage_account.backup_gb
  azure_files_share_name           = var.azure_files.share_name
  azure_files_quota_gb             = var.azure_files.quota_gb
  azure_files_zone_redundant       = var.azure_files.zone_redundant
  postgresql = {
    administrator_login           = var.postgresql_flexible_server.administrator_login
    zone                          = var.postgresql_flexible_server.zone
    read_replica_zone             = var.postgresql_flexible_server.read_replica_zone
    extensions                    = var.postgresql_flexible_server.extensions
    public_network_access_enabled = var.postgresql_flexible_server.public_network_access_enabled
    password_length               = var.postgresql_flexible_server.password_length
    password_special              = var.postgresql_flexible_server.password_special
  }
  redis_public_network_access_enabled = var.redis_cache.public_network_access_enabled
  managed_redis = {
    public_network_access = var.azure_managed_redis.public_network_access
    client_protocol       = var.azure_managed_redis.client_protocol
    eviction_policy       = var.azure_managed_redis.eviction_policy
  }
  storage = {
    account_tier             = var.storage_account.account_tier
    account_replication_type = var.storage_account.account_replication_type
    account_kind             = var.storage_account.account_kind
    public_network_access    = var.storage_account.public_network_access
    is_hns_enabled           = var.storage_account.is_hns_enabled
    delete_retention_days    = var.storage_account.delete_retention_days
  }
  azure_files = {
    public_network_access = var.azure_files.public_network_access
    enabled_protocol      = var.azure_files.enabled_protocol
  }
  acr = {
    admin_enabled                 = var.container_registry.admin_enabled
    public_network_access_enabled = var.container_registry.public_network_access_enabled
  }
  private_endpoint_connection = var.network.private_endpoint
  acr_sku                     = var.container_registry.sku
  private_endpoint_subnet_id  = local.core.private_endpoint_subnet_id
  vnet_id                     = local.core.vnet_id
  dns_resource_group_name     = local.core.resource_group_name
  foundry_storage             = local.foundry_storage_settings
}

data "azurerm_client_config" "current" {}
