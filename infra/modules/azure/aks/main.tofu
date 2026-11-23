locals {
  pools = {
    for k, v in var.node_pools : k => merge(v, { name = coalesce(v.name, k) })
  }
  system_pool_override = lookup(local.pools, "system", null)
  system_pool = local.system_pool_override != null ? {
    name            = local.system_pool_override.name
    vm_size         = local.system_pool_override.vm_size
    node_count      = local.system_pool_override.node_count
    os_disk_size_gb = local.system_pool_override.os_disk_size_gb
    } : {
    name            = var.system_pool.name
    vm_size         = var.system_pool.vm_size
    node_count      = var.system_pool.node_count
    os_disk_size_gb = var.system_pool.os_disk_size_gb
  }
  user_pools = {
    for k, v in local.pools : k => v if k != "system"
  }
  # Empty list = regional node pool (no zone pin). Required on some subs/regions (e.g. sandbox centralindia).
  node_pool_zones = length(var.availability_zones) > 0 ? var.availability_zones : null
}

resource "azurerm_kubernetes_cluster" "main" {
  name                    = "${var.name_prefix}-aks"
  location                = var.location
  resource_group_name     = var.resource_group_name
  dns_prefix              = var.name_prefix
  kubernetes_version      = var.kubernetes_version
  private_cluster_enabled = var.api_server.private_cluster_enabled
  private_dns_zone_id     = var.api_server.private_cluster_enabled ? var.api_server.private_dns_zone_id : null
  sku_tier                = var.cluster.sku_tier
  tags                    = var.tags

  default_node_pool {
    name            = local.system_pool.name
    vm_size         = local.system_pool.vm_size
    node_count      = local.system_pool.node_count
    vnet_subnet_id  = var.node_subnet_id
    os_disk_size_gb = local.system_pool.os_disk_size_gb
    os_disk_type    = var.os_disk_type
    zones           = local.node_pool_zones
  }

  identity {
    type = "SystemAssigned"
  }

  dynamic "api_server_access_profile" {
    for_each = length(var.api_server.authorized_ip_ranges) > 0 ? [1] : []
    content {
      authorized_ip_ranges = var.api_server.authorized_ip_ranges
    }
  }

  node_provisioning_profile {
    mode = var.cluster.node_provisioning_mode
  }

  network_profile {
    network_plugin      = var.network_profile.network_plugin
    network_plugin_mode = var.network_profile.network_plugin_mode == "none" ? null : var.network_profile.network_plugin_mode
    pod_cidr            = var.network_profile.network_plugin_mode == "overlay" || var.network_profile.network_plugin == "kubenet" ? var.pod_cidr : null
    service_cidr        = var.service_cidr
    dns_service_ip      = var.dns_service_ip
    outbound_type       = var.network_profile.outbound_type
  }

  oidc_issuer_enabled       = var.cluster.oidc_issuer_enabled
  workload_identity_enabled = var.cluster.workload_identity_enabled
}

resource "azurerm_kubernetes_cluster_node_pool" "user" {
  for_each = local.user_pools

  name                  = each.value.name
  kubernetes_cluster_id = azurerm_kubernetes_cluster.main.id
  vm_size               = each.value.vm_size
  node_count            = each.value.node_count
  vnet_subnet_id        = var.node_subnet_id
  os_disk_size_gb       = each.value.os_disk_size_gb
  os_disk_type          = var.os_disk_type
  mode                  = each.value.mode
  zones                 = local.node_pool_zones
}
