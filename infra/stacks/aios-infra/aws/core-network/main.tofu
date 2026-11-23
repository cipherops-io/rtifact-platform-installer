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

locals {
  bastion_key_file = "${var.name_prefix}-bastion.pem"
  bastion_ssh_command = length(module.bastion) == 0 ? null : (
    module.bastion[0].access == "ssm"
    ? "aws ssm start-session --target ${module.bastion[0].instance_id} --region ${var.location}"
    : join(" ", compact([
      "ssh",
      try(var.bastion.ssh_public_key, "") == "" ? "-i ${local.bastion_key_file}" : "",
      module.bastion[0].ssh_port != 22 ? "-p ${module.bastion[0].ssh_port}" : "",
      "${module.bastion[0].admin_username}@${coalesce(module.bastion[0].public_ip, module.bastion[0].private_ip)}",
    ]))
  )
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