resource "azurerm_redis_cache" "main" {
  count = var.features.redis_cache ? 1 : 0

  name                          = "${var.name_prefix}-redis"
  location                      = var.location
  resource_group_name           = var.resource_group_name
  capacity                      = var.redis_capacity
  family                        = var.redis_family
  sku_name                      = var.redis_sku_tier
  minimum_tls_version           = var.redis_minimum_tls_version
  tags                          = var.tags
  public_network_access_enabled = false
}