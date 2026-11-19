resource "random_password" "pg" {
  count = var.features.postgresql_flexible_server ? 1 : 0

  length  = 24
  special = true
}

resource "azurerm_postgresql_flexible_server" "main" {
  count = var.features.postgresql_flexible_server ? 1 : 0

  name                   = "${var.name_prefix}-pg"
  resource_group_name    = var.resource_group_name
  location               = var.location
  version                = var.postgresql_version
  administrator_login    = "aiosadmin"
  administrator_password = random_password.pg[0].result
  sku_name               = var.postgresql_sku_name
  storage_mb             = var.postgresql_storage_mb
  zone                   = "1"
  tags                   = var.tags

  public_network_access_enabled = false
  backup_retention_days         = var.postgresql_backup_retention_days

  dynamic "high_availability" {
    for_each = var.postgresql_ha_enabled ? [1] : []
    content {
      mode = var.postgresql_ha_mode
    }
  }
}

resource "azurerm_postgresql_flexible_server" "read_replica" {
  count = var.features.postgresql_flexible_server && var.postgresql_read_replica_enabled ? 1 : 0

  name                = "${var.name_prefix}-pg-ro"
  resource_group_name = var.resource_group_name
  location            = var.location
  create_mode         = "Replica"
  source_server_id    = azurerm_postgresql_flexible_server.main[0].id
  zone                = "2"
  tags                = var.tags

  public_network_access_enabled = false
}

resource "azurerm_postgresql_flexible_server_configuration" "extensions" {
  count = var.features.postgresql_flexible_server ? 1 : 0

  name      = "azure.extensions"
  server_id = azurerm_postgresql_flexible_server.main[0].id
  value     = "CITEXT,UUID-OSSP"
}
