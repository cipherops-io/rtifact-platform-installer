resource "google_kms_key_ring" "main" {
  name     = "${var.name_prefix}-kr"
  location = var.location
}

resource "google_kms_crypto_key" "main" {
  name     = "${var.name_prefix}-key"
  key_ring = google_kms_key_ring.main.id
}
