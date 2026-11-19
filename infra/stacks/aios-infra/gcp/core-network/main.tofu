module "naming" {
  count = var.features.resource_group ? 1 : 0

  source     = "../../../../modules/gcp/naming"
  prefix     = var.name_prefix
  location   = var.location
  project_id = var.project_id
}

module "network" {
  count = var.features.virtual_network ? 1 : 0

  source                       = "../../../../modules/gcp/vpc"
  project_id                   = var.project_id
  scope_name                   = module.naming[0].scope_name
  location                     = var.location
  name_prefix                  = var.name_prefix
  tags                         = module.naming[0].tags
  vnet_cidr                    = var.network.vnet_cidr
  node_subnet_cidr             = var.network.node_subnet_cidr
  load_balancer_subnet_cidr    = var.network.load_balancer_subnet_cidr
  private_endpoint_subnet_cidr = var.network.private_endpoint_subnet_cidr
  management_subnet_cidr       = var.network.management_subnet_cidr
}
