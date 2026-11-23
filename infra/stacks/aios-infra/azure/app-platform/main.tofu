data "terraform_remote_state" "core_network" {
  backend = var.upstream_state["core-network"].backend
  config  = var.upstream_state["core-network"].config
}

locals {
  core = data.terraform_remote_state.core_network.outputs
}

module "aks" {
  count = var.features.kubernetes_cluster ? 1 : 0

  source              = "../../../../modules/azure/aks"
  resource_group_name = local.core.resource_group_name
  location            = var.location
  name_prefix         = var.name_prefix
  tags                = local.core.tags
  kubernetes_version  = var.kubernetes_cluster.kubernetes_version
  node_subnet_id      = local.core.node_subnet_id
  pod_cidr            = var.network.pod_cidr
  service_cidr        = var.network.service_cidr
  dns_service_ip      = var.network.dns_service_ip
  node_pools          = var.kubernetes_cluster.node_pools
  availability_zones  = var.kubernetes_cluster.availability_zones
  os_disk_type        = var.kubernetes_cluster.os_disk_type
  system_pool         = var.kubernetes_cluster.system_pool
  api_server          = var.kubernetes_cluster.api_server
  network_profile     = var.kubernetes_cluster.network_profile
  cluster = {
    sku_tier                  = var.kubernetes_cluster.sku_tier
    oidc_issuer_enabled       = var.kubernetes_cluster.oidc_issuer_enabled
    workload_identity_enabled = var.kubernetes_cluster.workload_identity_enabled
    node_provisioning_mode    = var.kubernetes_cluster.node_provisioning_mode
  }
}

module "compute" {
  count = var.features.azure_bastion ? 1 : 0

  source              = "../../../../modules/azure/compute"
  resource_group_name = local.core.resource_group_name
  location            = var.location
  name_prefix         = var.name_prefix
  tags                = local.core.tags
  subnet_id           = local.core.management_subnet_id
  access              = var.bastion.access
  allowed_ssh_cidrs   = var.bastion.allowed_ssh_cidrs
  vm_size             = var.bastion.vm_size
  os_disk_size_gb     = var.bastion.os_disk_size_gb
  os_disk_type        = var.bastion.os_disk_type
  os_disk_caching     = var.bastion.os_disk_caching
  admin_username      = var.bastion.admin_username
  image               = var.bastion.image
  ssh_key             = var.bastion.ssh_key
  private_ip          = var.bastion.private_ip
}

locals {
  bastion_key_file = "${var.name_prefix}-bastion.pem"
  bastion          = length(module.compute) > 0 ? module.compute[0] : null
  bastion_ssh_command = local.bastion == null ? null : (
    var.bastion.access == "azure_bastion" ? join(" ", [
      "az network bastion ssh",
      "--subscription ${var.subscription_id}",
      "--name ${try(local.core.azure_bastion_name, "")}",
      "--resource-group ${local.core.resource_group_name}",
      "--target-resource-id ${local.bastion.bastion_vm_id}",
      "--auth-type ssh-key",
      "--username ${local.bastion.bastion_admin_username}",
      "--ssh-key ${local.bastion_key_file}",
    ]) :
    "ssh -i ${local.bastion_key_file} ${local.bastion.bastion_admin_username}@${var.bastion.access == "public_ip" ? local.bastion.bastion_public_ip : local.bastion.bastion_private_ip}"
  )
}
