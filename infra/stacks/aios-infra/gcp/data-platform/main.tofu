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
    var.features.redis_cache
  )
}

module "security" {
  count = var.features.key_vault ? 1 : 0

  source      = "../../../../modules/gcp/security"
  scope_name  = local.core.resource_group_name
  location    = var.location
  name_prefix = var.name_prefix
  tags        = local.core.tags
  sku         = var.key_vault.sku
}

module "data" {
  count = local.data_services_enabled ? 1 : 0

  source                     = "../../../../modules/gcp/data"
  project_id                 = var.project_id
  scope_name                 = local.core.resource_group_name
  location                   = var.location
  name_prefix                = var.name_prefix
  tags                       = local.core.tags
  features                   = var.features
  postgresql_version         = var.postgresql_flexible_server.version
  postgresql_sku_name        = var.postgresql_flexible_server.sku_name
  postgresql_storage_mb      = var.postgresql_flexible_server.storage_mb
  redis_sku_name             = var.redis_cache.sku_name
  storage_corpus_gb          = var.storage_account.corpus_gb
  storage_backup_gb          = var.storage_account.backup_gb
  acr_sku                    = var.container_registry.sku
  private_endpoint_subnet_id = local.core.private_endpoint_subnet_id
  node_subnet_id             = local.core.node_subnet_id
  vnet_id                    = local.core.vnet_id
  dns_scope_name             = local.core.resource_group_name
}
