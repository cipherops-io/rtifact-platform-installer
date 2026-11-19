resource "google_compute_network" "main" {
  name                    = "${var.name_prefix}-vpc"
  auto_create_subnetworks = false
  routing_mode            = "REGIONAL"
}

resource "google_compute_subnetwork" "node" {
  name                     = "aks-nodes"
  ip_cidr_range            = var.node_subnet_cidr
  region                   = var.location
  network                  = google_compute_network.main.id
  private_ip_google_access = true
}

resource "google_compute_subnetwork" "load_balancer" {
  name          = "load-balancer"
  ip_cidr_range = var.load_balancer_subnet_cidr
  region        = var.location
  network       = google_compute_network.main.id
}

resource "google_compute_subnetwork" "private_endpoint" {
  name                     = "private-endpoints"
  ip_cidr_range            = var.private_endpoint_subnet_cidr
  region                   = var.location
  network                  = google_compute_network.main.id
  private_ip_google_access = true
}

resource "google_compute_subnetwork" "management" {
  name                     = "management"
  ip_cidr_range            = var.management_subnet_cidr
  region                   = var.location
  network                  = google_compute_network.main.id
  private_ip_google_access = true
}

resource "google_compute_router" "main" {
  name    = "${var.name_prefix}-router"
  region  = var.location
  network = google_compute_network.main.id
}

resource "google_compute_router_nat" "main" {
  name                               = "${var.name_prefix}-nat"
  router                             = google_compute_router.main.name
  region                             = var.location
  nat_ip_allocate_option             = "AUTO_ONLY"
  source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"
}
