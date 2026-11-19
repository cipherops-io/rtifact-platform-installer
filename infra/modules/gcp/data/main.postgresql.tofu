resource "random_password" "pg" {
  count   = var.features.postgresql_flexible_server ? 1 : 0
  length  = 24
  special = true
}

resource "google_sql_database_instance" "postgresql" {
  count = var.features.postgresql_flexible_server ? 1 : 0

  name             = "${var.name_prefix}-pg"
  database_version = var.postgresql_version
  region           = var.location

  settings {
    tier      = var.postgresql_sku_name
    disk_size = max(floor(var.postgresql_storage_mb / 1024), 20)

    ip_configuration {
      ipv4_enabled    = false
      private_network = var.vnet_id
    }
  }

  deletion_protection = false
}