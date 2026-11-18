locals {
  data_services_enabled = (
    var.features.postgresql_flexible_server ||
    var.features.container_registry ||
    var.features.storage_account ||
    var.features.redis_cache
  )
  vnet_id                    = try(var.layer_inputs.vnet_id, module.network[0].vnet_id)
  private_endpoint_subnet_id = try(var.layer_inputs.private_endpoint_subnet_id, module.network[0].private_endpoint_subnet_id)
  use_embedded_network       = length(var.layer_inputs) == 0
}

module "naming" {
  count = var.features.resource_group ? 1 : 0
  source   = "../../../modules/naming"
  prefix   = var.name_prefix
  location = var.location
}

module "network" {
  count = local.use_embedded_network && var.features.virtual_network ? 1 : 0
  source                       = "../../../modules/network"
  resource_group_name          = module.naming[0].resource_group_name
  location                     = var.location
  name_prefix                  = var.name_prefix
  tags                         = module.naming[0].tags
  vnet_cidr                    = var.network.vnet_cidr
  node_subnet_cidr             = var.network.node_subnet_cidr
  load_balancer_subnet_cidr    = var.network.load_balancer_subnet_cidr
  private_endpoint_subnet_cidr = var.network.private_endpoint_subnet_cidr
  management_subnet_cidr       = var.network.management_subnet_cidr
}

module "security" {
  count = var.features.key_vault ? 1 : 0
  source              = "../../../modules/security"
  resource_group_name = module.naming[0].resource_group_name
  location            = var.location
  name_prefix         = var.name_prefix
  tags                = module.naming[0].tags
  sku                 = var.key_vault.sku
  tenant_id           = data.azurerm_client_config.current.tenant_id
}

module "data" {
  count = local.data_services_enabled ? 1 : 0
  source                     = "../../../modules/data"
  resource_group_name        = module.naming[0].resource_group_name
  location                   = var.location
  name_prefix                = var.name_prefix
  tags                       = module.naming[0].tags
  features                   = var.features
  postgresql_version         = var.postgresql_flexible_server.version
  postgresql_sku_name        = var.postgresql_flexible_server.sku_name
  postgresql_storage_mb      = var.postgresql_flexible_server.storage_mb
  redis_sku_name             = var.redis_cache.sku_name
  storage_corpus_gb          = var.storage_account.corpus_gb
  storage_backup_gb          = var.storage_account.backup_gb
  acr_sku                    = var.container_registry.sku
  private_endpoint_subnet_id = local.private_endpoint_subnet_id
  vnet_id                    = local.vnet_id
  dns_resource_group_name    = module.naming[0].resource_group_name
}

data "azurerm_client_config" "current" {}