locals {
  node_subnet_id       = try(var.layer_inputs.node_subnet_id, module.network[0].node_subnet_id)
  management_subnet_id = try(var.layer_inputs.management_subnet_id, module.network[0].management_subnet_id)
  use_embedded_network = length(var.layer_inputs) == 0
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

module "aks" {
  count = var.features.kubernetes_cluster ? 1 : 0
  source              = "../../../modules/aks"
  resource_group_name = module.naming[0].resource_group_name
  location            = var.location
  name_prefix         = var.name_prefix
  tags                = module.naming[0].tags
  kubernetes_version  = var.kubernetes_cluster.kubernetes_version
  node_subnet_id      = local.node_subnet_id
  pod_cidr            = var.network.pod_cidr
  service_cidr        = var.network.service_cidr
  dns_service_ip      = var.network.dns_service_ip
  node_pools          = var.kubernetes_cluster.node_pools
  availability_zones  = var.kubernetes_cluster.availability_zones
}

module "compute" {
  count = var.features.azure_bastion ? 1 : 0
  source              = "../../../modules/compute"
  resource_group_name = module.naming[0].resource_group_name
  location            = var.location
  name_prefix         = var.name_prefix
  tags                = module.naming[0].tags
  subnet_id           = local.management_subnet_id
  vm_size             = var.azure_bastion.vm_size
  os_disk_size_gb     = var.azure_bastion.os_disk_size_gb
}

data "azurerm_client_config" "current" {}