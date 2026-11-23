resource "azurerm_managed_redis" "main" {
  count = var.features.azure_managed_redis ? 1 : 0

  name                  = "${var.name_prefix}-mredis"
  resource_group_name   = var.resource_group_name
  location              = var.location
  sku_name              = var.managed_redis_sku_name
  public_network_access = var.managed_redis.public_network_access
  tags                  = var.tags

  default_database {
    client_protocol = var.managed_redis.client_protocol
    eviction_policy = var.managed_redis.eviction_policy
  }
}