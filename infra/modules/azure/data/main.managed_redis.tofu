resource "azurerm_managed_redis" "main" {
  count = var.features.azure_managed_redis ? 1 : 0

  name                  = "${var.name_prefix}-mredis"
  resource_group_name   = var.resource_group_name
  location              = var.location
  sku_name              = var.managed_redis_sku_name
  public_network_access = "Disabled"
  tags                  = var.tags

  default_database {
    client_protocol = "Encrypted"
    eviction_policy = "AllKeysLRU"
  }
}