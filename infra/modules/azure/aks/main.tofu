locals {
  system_pool = lookup(var.node_pools, "system", {
    name       = "system"
    vm_size    = "Standard_D4s_v5"
    node_count = 3
  })
  user_pools = {
    for k, v in var.node_pools : k => v if k != "system"
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
  private_cluster_enabled = true
  private_dns_zone_id     = "System"
  sku_tier                = "Standard"
  tags                    = var.tags

  default_node_pool {
    name            = local.system_pool.name
    vm_size         = local.system_pool.vm_size
    node_count      = local.system_pool.node_count
    vnet_subnet_id  = var.node_subnet_id
    os_disk_size_gb = 128
    os_disk_type    = var.os_disk_type
    zones           = local.node_pool_zones
  }

  identity {
    type = "SystemAssigned"
  }

  node_provisioning_profile {
    mode = "Manual"
  }

  network_profile {
    network_plugin      = "azure"
    network_plugin_mode = "overlay"
    pod_cidr            = var.pod_cidr
    service_cidr        = var.service_cidr
    dns_service_ip      = var.dns_service_ip
    outbound_type       = "userAssignedNATGateway"
  }

  oidc_issuer_enabled       = true
  workload_identity_enabled = true
}

resource "azurerm_kubernetes_cluster_node_pool" "user" {
  for_each = local.user_pools

  name                  = each.value.name
  kubernetes_cluster_id = azurerm_kubernetes_cluster.main.id
  vm_size               = each.value.vm_size
  node_count            = each.value.node_count
  vnet_subnet_id        = var.node_subnet_id
  os_disk_size_gb       = 128
  os_disk_type          = var.os_disk_type
  mode                  = "User"
  zones                 = local.node_pool_zones
}