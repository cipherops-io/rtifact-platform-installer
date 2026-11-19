resource "google_compute_forwarding_rule" "psc" {
  name                  = var.name
  region                = var.location
  network               = var.network_id
  load_balancing_scheme = ""
  ip_address            = null
  target                = var.target_service
}