locals {
  postgres_dns = var.features.private_dns_zones && var.features.postgresql_flexible_server
  redis_dns    = var.features.private_dns_zones && var.features.redis_cache
  blob_dns     = var.features.private_dns_zones && var.features.storage_account
  acr_dns      = var.features.private_dns_zones && var.features.container_registry
}

resource "azurerm_private_dns_zone" "postgres" {
  count = local.postgres_dns ? 1 : 0

  name                = "privatelink.postgres.database.azure.com"
  resource_group_name = var.dns_resource_group_name
  tags                = var.tags
}

resource "azurerm_private_dns_zone_virtual_network_link" "postgres" {
  count = local.postgres_dns ? 1 : 0

  name                  = "${var.name_prefix}-pg-dns"
  private_dns_zone_id   = azurerm_private_dns_zone.postgres[0].id
  virtual_network_id    = var.vnet_id
}

resource "azurerm_private_dns_zone" "redis" {
  count = local.redis_dns ? 1 : 0

  name                = "privatelink.redis.cache.windows.net"
  resource_group_name = var.dns_resource_group_name
  tags                = var.tags
}

resource "azurerm_private_dns_zone_virtual_network_link" "redis" {
  count = local.redis_dns ? 1 : 0

  name                  = "${var.name_prefix}-redis-dns"
  private_dns_zone_id   = azurerm_private_dns_zone.redis[0].id
  virtual_network_id    = var.vnet_id
}

resource "azurerm_private_dns_zone" "blob" {
  count = local.blob_dns ? 1 : 0

  name                = "privatelink.blob.core.windows.net"
  resource_group_name = var.dns_resource_group_name
  tags                = var.tags
}

resource "azurerm_private_dns_zone_virtual_network_link" "blob" {
  count = local.blob_dns ? 1 : 0

  name                = "${var.name_prefix}-blob-dns"
  private_dns_zone_id = azurerm_private_dns_zone.blob[0].id
  virtual_network_id  = var.vnet_id
}

resource "azurerm_private_dns_zone" "acr" {
  count = local.acr_dns ? 1 : 0

  name                = "privatelink.azurecr.io"
  resource_group_name = var.dns_resource_group_name
  tags                = var.tags
}

resource "azurerm_private_dns_zone_virtual_network_link" "acr" {
  count = local.acr_dns ? 1 : 0

  name                = "${var.name_prefix}-acr-dns"
  private_dns_zone_id = azurerm_private_dns_zone.acr[0].id
  virtual_network_id  = var.vnet_id
}