resource "azurerm_redis_cache" "main" {
  count = var.features.redis_cache ? 1 : 0

  name                          = "${var.name_prefix}-redis"
  location                      = var.location
  resource_group_name           = var.resource_group_name
  capacity                      = 1
  family                        = "P"
  sku_name                      = "Premium"
  minimum_tls_version           = "1.2"
  tags                          = var.tags
  public_network_access_enabled = false
}