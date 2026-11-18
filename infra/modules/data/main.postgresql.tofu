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
}

resource "azurerm_postgresql_flexible_server_configuration" "extensions" {
  count = var.features.postgresql_flexible_server ? 1 : 0

  name      = "azure.extensions"
  server_id = azurerm_postgresql_flexible_server.main[0].id
  value     = "CITEXT,UUID-OSSP"
}