module "naming" {
  count = var.features.network ? 1 : 0

  source   = "../../../../modules/aws/naming"
  prefix   = var.name_prefix
  location = var.location
}

locals {
  base_tags          = try(var.tags, {})
  net_tags           = var.features.network ? merge(local.base_tags, module.naming[0].tags) : local.base_tags
  cluster_name       = try(var.kubernetes_cluster.name, "${var.name_prefix}-eks")
  public_subnet_tier = try(var.bastion.subnet_tier, "public")
}

module "network" {
  count = var.features.network ? 1 : 0

  source       = "../../../../modules/aws/vpc"
  name_prefix  = var.name_prefix
  location     = var.location
  tags         = local.net_tags
  cluster_name = local.cluster_name
  network      = var.network
}

module "bastion" {
  count = var.features.bastion ? 1 : 0

  source      = "../../../../modules/aws/bastion"
  name_prefix = var.name_prefix
  location    = var.location
  tags        = local.net_tags
  vpc_id      = module.network[0].vpc_id
  subnet_id   = module.network[0].subnet_ids_by_tier[local.public_subnet_tier][0]
  bastion     = var.bastion

  depends_on = [module.network]
}