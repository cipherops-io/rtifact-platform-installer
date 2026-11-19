output "postgresql_id" {
  value = try(azurerm_postgresql_flexible_server.main[0].id, null)
}

output "storage_account_id" {
  value = try(azurerm_storage_account.main[0].id, null)
}

output "acr_id" {
  value = try(azurerm_container_registry.main[0].id, null)
}

output "redis_id" {
  value = try(azurerm_redis_cache.main[0].id, null)
}

output "azure_files_storage_account_id" {
  value = try(azurerm_storage_account.files[0].id, null)
}

output "azure_files_share_name" {
  value = try(azurerm_storage_share.aios[0].name, null)
}

output "private_dns_zone_ids" {
  value = {
    postgres = try(azurerm_private_dns_zone.postgres[0].id, null)
    redis    = try(azurerm_private_dns_zone.redis[0].id, null)
    blob     = try(azurerm_private_dns_zone.blob[0].id, null)
    acr      = try(azurerm_private_dns_zone.acr[0].id, null)
  }
}