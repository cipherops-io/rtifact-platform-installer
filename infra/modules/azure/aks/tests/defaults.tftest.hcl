mock_provider "azurerm" {
  mock_resource "azurerm_kubernetes_cluster" {
    defaults = {
      id = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.ContainerService/managedClusters/t-aks"
    }
  }
}

variables {
  resource_group_name = "rg"
  location            = "centralindia"
  name_prefix         = "t"
  kubernetes_version  = "1.33"
  node_subnet_id      = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/v/subnets/s"
  pod_cidr            = "10.244.0.0/16"
  service_cidr        = "10.0.0.0/16"
  dns_service_ip      = "10.0.0.10"
  node_pools = {
    apps = { vm_size = "Standard_D4s_v5", node_count = 2 }
  }
}

run "defaults_match_previous_literals" {
  command = plan

  assert {
    condition = (
      azurerm_kubernetes_cluster.main.private_cluster_enabled &&
      azurerm_kubernetes_cluster.main.private_dns_zone_id == "System" &&
      azurerm_kubernetes_cluster.main.sku_tier == "Standard" &&
      azurerm_kubernetes_cluster.main.oidc_issuer_enabled &&
      azurerm_kubernetes_cluster.main.workload_identity_enabled
    )
    error_message = "cluster defaults changed"
  }
  assert {
    condition = (
      azurerm_kubernetes_cluster.main.default_node_pool[0].name == "system" &&
      azurerm_kubernetes_cluster.main.default_node_pool[0].vm_size == "Standard_D4s_v5" &&
      azurerm_kubernetes_cluster.main.default_node_pool[0].node_count == 3 &&
      azurerm_kubernetes_cluster.main.default_node_pool[0].os_disk_size_gb == 128
    )
    error_message = "system pool defaults changed"
  }
  assert {
    condition = (
      azurerm_kubernetes_cluster.main.network_profile[0].network_plugin == "azure" &&
      azurerm_kubernetes_cluster.main.network_profile[0].network_plugin_mode == "overlay" &&
      azurerm_kubernetes_cluster.main.network_profile[0].outbound_type == "userAssignedNATGateway" &&
      azurerm_kubernetes_cluster.main.network_profile[0].pod_cidr == "10.244.0.0/16"
    )
    error_message = "network profile defaults changed"
  }
  assert {
    condition     = azurerm_kubernetes_cluster.main.node_provisioning_profile[0].mode == "Manual"
    error_message = "node provisioning default changed"
  }
  assert {
    condition = (
      azurerm_kubernetes_cluster_node_pool.user["apps"].name == "apps" &&
      azurerm_kubernetes_cluster_node_pool.user["apps"].mode == "User" &&
      azurerm_kubernetes_cluster_node_pool.user["apps"].os_disk_size_gb == 128
    )
    error_message = "user pool defaults changed (name falls back to the map key)"
  }
}

run "rejects_authorized_ranges_on_private_cluster" {
  command = plan

  variables {
    api_server = { authorized_ip_ranges = ["203.0.113.0/24"] }
  }

  expect_failures = [var.api_server]
}
