data "terraform_remote_state" "core_network" {
  backend = var.upstream_state["core-network"].backend
  config  = var.upstream_state["core-network"].config
}

locals {
  core = data.terraform_remote_state.core_network.outputs
}

module "eks" {
  count = var.features.kubernetes_cluster ? 1 : 0

  source             = "../../../../modules/aws/eks"
  scope_name         = local.core.resource_group_name
  location           = var.location
  name_prefix        = var.name_prefix
  tags               = local.core.tags
  kubernetes_version = var.kubernetes_cluster.kubernetes_version
  node_subnet_id     = local.core.node_subnet_id
  pod_cidr           = var.network.pod_cidr
  service_cidr       = var.network.service_cidr
  dns_service_ip     = var.network.dns_service_ip
  node_pools         = var.kubernetes_cluster.node_pools
  availability_zones = var.kubernetes_cluster.availability_zones
}

module "compute" {
  count = var.features.azure_bastion ? 1 : 0

  source          = "../../../../modules/aws/compute"
  scope_name      = local.core.resource_group_name
  location        = var.location
  name_prefix     = var.name_prefix
  tags            = local.core.tags
  subnet_id       = local.core.management_subnet_id
  vm_size         = var.azure_bastion.vm_size
  os_disk_size_gb = var.azure_bastion.os_disk_size_gb
}
