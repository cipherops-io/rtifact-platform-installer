locals {
  system_pool = lookup(var.node_pools, "system", {
    name       = "system"
    vm_size    = "e2-medium"
    node_count = 2
  })
}

resource "google_container_cluster" "main" {
  name     = "${var.name_prefix}-gke"
  location = var.location

  remove_default_node_pool = true
  initial_node_count       = 1

  network    = var.network_id
  subnetwork = var.node_subnet_id

  private_cluster_config {
    enable_private_nodes    = true
    enable_private_endpoint = true
    master_ipv4_cidr_block  = "172.16.0.0/28"
  }
}

resource "google_container_node_pool" "system" {
  name       = local.system_pool.name
  location   = var.location
  cluster    = google_container_cluster.main.name
  node_count = local.system_pool.node_count

  node_config {
    machine_type = local.system_pool.vm_size
    oauth_scopes = ["https://www.googleapis.com/auth/cloud-platform"]
  }
}