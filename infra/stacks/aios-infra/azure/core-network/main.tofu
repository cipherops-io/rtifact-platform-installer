# core-network is the only layer that creates the naming/resource group.
module "naming" {
  count = var.features.resource_group ? 1 : 0

  source                = "../../../../modules/azure/naming"
  prefix                = var.name_prefix
  location              = var.location
  create_resource_group = var.resource_group_create
}

locals {
  azure_bastion_host = var.features.azure_bastion && var.bastion.access == "azure_bastion"
  net_tags           = var.features.resource_group ? module.naming[0].tags : {}
  net_rg             = var.features.resource_group ? module.naming[0].resource_group_name : null
}

module "network" {
  count = var.features.virtual_network ? 1 : 0

  source                            = "../../../../modules/azure/network"
  resource_group_name               = local.net_rg
  location                          = var.location
  name_prefix                       = var.name_prefix
  tags                              = local.net_tags
  vnet_cidr                         = var.network.vnet_cidr
  node_subnet_cidr                  = var.network.node_subnet_cidr
  load_balancer_subnet_cidr         = var.network.load_balancer_subnet_cidr
  private_endpoint_subnet_cidr      = var.network.private_endpoint_subnet_cidr
  management_subnet_cidr            = var.network.management_subnet_cidr
  features                          = var.features
  azure_firewall_subnet_cidr        = try(var.network.azure_firewall_subnet_cidr, null)
  gateway_subnet_cidr               = try(var.network.gateway_subnet_cidr, null)
  azure_bastion_subnet_cidr         = try(var.network.azure_bastion_subnet_cidr, null)
  dns_resolver_inbound_subnet_cidr  = try(var.network.dns_resolver_inbound_subnet_cidr, null)
  dns_resolver_outbound_subnet_cidr = try(var.network.dns_resolver_outbound_subnet_cidr, null)
  bastion_managed                   = local.azure_bastion_host
  nat                               = var.network.nat
  dns_resolver_outbound_enabled     = var.private_dns_resolver.outbound_endpoint_enabled
  dns_resolver_inbound = {
    enabled                      = var.private_dns_resolver.inbound_endpoint_enabled
    private_ip_allocation_method = var.private_dns_resolver.inbound_private_ip_allocation_method
    private_ip_address           = var.private_dns_resolver.inbound_private_ip_address
  }
}

module "firewall" {
  count = var.features.azure_firewall && var.features.virtual_network && try(var.network.azure_firewall_subnet_cidr, null) != null ? 1 : 0

  source              = "../../../../modules/azure/firewall"
  resource_group_name = local.net_rg
  location            = var.location
  name_prefix         = var.name_prefix
  tags                = local.net_tags
  firewall_subnet_id  = module.network[0].firewall_subnet_id
  sku_tier            = var.azure_firewall.sku_tier

  depends_on = [module.network]
}

module "connectivity" {
  count = var.features.virtual_network && try(var.network.gateway_subnet_cidr, null) != null && (var.features.azure_vpn_gateway || var.features.azure_express_route_gateway) ? 1 : 0

  source                = "../../../../modules/azure/connectivity"
  resource_group_name   = local.net_rg
  location              = var.location
  name_prefix           = var.name_prefix
  tags                  = local.net_tags
  gateway_subnet_id     = module.network[0].gateway_subnet_id
  vpn_enabled           = var.features.azure_vpn_gateway
  express_route_enabled = var.features.azure_express_route_gateway
  vpn                   = var.azure_vpn
  express_route         = var.azure_express_route

  depends_on = [module.network]
}

module "bastion_host" {
  count = local.azure_bastion_host && var.features.virtual_network && try(var.network.azure_bastion_subnet_cidr, null) != null ? 1 : 0

  source              = "../../../../modules/azure/bastion-host"
  resource_group_name = local.net_rg
  location            = var.location
  name_prefix         = var.name_prefix
  tags                = local.net_tags
  bastion_subnet_id   = module.network[0].bastion_subnet_id
  settings            = var.bastion.azure_bastion

  depends_on = [module.network]
}