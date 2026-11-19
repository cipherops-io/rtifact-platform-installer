resource "azurerm_private_endpoint" "postgres" {
  count = var.features.postgresql_flexible_server ? 1 : 0

  name                = "${var.name_prefix}-pg-pe"
  location            = var.location
  resource_group_name = var.resource_group_name
  subnet_id           = var.private_endpoint_subnet_id
  tags                = var.tags

  private_service_connection {
    name                           = "postgres"
    private_connection_resource_id = azurerm_postgresql_flexible_server.main[0].id
    subresource_names              = ["postgresqlServer"]
    is_manual_connection           = false
  }

  dynamic "private_dns_zone_group" {
    for_each = local.postgres_dns ? [1] : []
    content {
      name                 = "default"
      private_dns_zone_ids = [azurerm_private_dns_zone.postgres[0].id]
    }
  }
}

resource "azurerm_private_endpoint" "redis" {
  count = var.features.redis_cache ? 1 : 0

  name                = "${var.name_prefix}-redis-pe"
  location            = var.location
  resource_group_name = var.resource_group_name
  subnet_id           = var.private_endpoint_subnet_id
  tags                = var.tags

  private_service_connection {
    name                           = "redis"
    private_connection_resource_id = azurerm_redis_cache.main[0].id
    subresource_names              = ["redisCache"]
    is_manual_connection           = false
  }

  dynamic "private_dns_zone_group" {
    for_each = local.redis_dns ? [1] : []
    content {
      name                 = "default"
      private_dns_zone_ids = [azurerm_private_dns_zone.redis[0].id]
    }
  }
}

resource "azurerm_private_endpoint" "blob" {
  count = var.features.storage_account ? 1 : 0

  name                = "${var.name_prefix}-blob-pe"
  location            = var.location
  resource_group_name = var.resource_group_name
  subnet_id           = var.private_endpoint_subnet_id
  tags                = var.tags

  private_service_connection {
    name                           = "blob"
    private_connection_resource_id = azurerm_storage_account.main[0].id
    subresource_names              = ["blob"]
    is_manual_connection           = false
  }

  dynamic "private_dns_zone_group" {
    for_each = local.blob_dns ? [1] : []
    content {
      name                 = "default"
      private_dns_zone_ids = [azurerm_private_dns_zone.blob[0].id]
    }
  }
}

resource "azurerm_private_endpoint" "acr" {
  count = var.features.container_registry ? 1 : 0

  name                = "${var.name_prefix}-acr-pe"
  location            = var.location
  resource_group_name = var.resource_group_name
  subnet_id           = var.private_endpoint_subnet_id
  tags                = var.tags

  private_service_connection {
    name                           = "acr"
    private_connection_resource_id = azurerm_container_registry.main[0].id
    subresource_names              = ["registry"]
    is_manual_connection           = false
  }

  dynamic "private_dns_zone_group" {
    for_each = local.acr_dns ? [1] : []
    content {
      name                 = "default"
      private_dns_zone_ids = [azurerm_private_dns_zone.acr[0].id]
    }
  }
}