locals {
  data_services_enabled = (
    var.features.postgresql_flexible_server ||
    var.features.container_registry ||
    var.features.storage_account ||
    var.features.redis_cache
  )
}

module "naming" {
  count = var.features.resource_group ? 1 : 0

  source   = "../../modules/naming"
  prefix   = var.name_prefix
  location = var.location
}

module "network" {
  count = var.features.virtual_network ? 1 : 0

  source                       = "../../modules/network"
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

  source              = "../../modules/security"
  resource_group_name = module.naming[0].resource_group_name
  location            = var.location
  name_prefix         = var.name_prefix
  tags                = module.naming[0].tags
  sku                 = var.key_vault.sku
  tenant_id           = data.azurerm_client_config.current.tenant_id
}

module "data" {
  count = local.data_services_enabled ? 1 : 0

  source                     = "../../modules/data"
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
  private_endpoint_subnet_id = module.network[0].private_endpoint_subnet_id
  vnet_id                    = module.network[0].vnet_id
  dns_resource_group_name    = module.naming[0].resource_group_name
}

module "aks" {
  count = var.features.kubernetes_cluster ? 1 : 0

  source             = "../../modules/aks"
  resource_group_name = module.naming[0].resource_group_name
  location           = var.location
  name_prefix        = var.name_prefix
  tags               = module.naming[0].tags
  kubernetes_version = var.kubernetes_cluster.kubernetes_version
  node_subnet_id     = module.network[0].node_subnet_id
  pod_cidr           = var.network.pod_cidr
  service_cidr       = var.network.service_cidr
  dns_service_ip     = var.network.dns_service_ip
  node_pools         = var.kubernetes_cluster.node_pools
  availability_zones = var.kubernetes_cluster.availability_zones

  depends_on = [module.network]
}

module "compute" {
  count = var.features.azure_bastion ? 1 : 0

  source              = "../../modules/compute"
  resource_group_name = module.naming[0].resource_group_name
  location            = var.location
  name_prefix         = var.name_prefix
  tags                = module.naming[0].tags
  subnet_id           = module.network[0].management_subnet_id
  vm_size             = var.azure_bastion.vm_size
  os_disk_size_gb     = var.azure_bastion.os_disk_size_gb
}

data "azurerm_client_config" "current" {}