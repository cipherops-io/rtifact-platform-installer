resource "google_secret_manager_secret" "workload" {
  count = var.features.postgresql_flexible_server ? 1 : 0

  secret_id = "${var.name_prefix}-workload-secrets"

  replication {
    auto {}
  }
}