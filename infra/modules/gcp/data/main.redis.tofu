resource "google_redis_instance" "main" {
  count = var.features.redis_cache ? 1 : 0

  name               = "${var.name_prefix}-redis"
  tier               = startswith(var.redis_sku_name, "STANDARD") ? "STANDARD_HA" : "BASIC"
  memory_size_gb     = 1
  region             = var.location
  authorized_network = var.vnet_id
  connect_mode       = "PRIVATE_SERVICE_ACCESS"
}