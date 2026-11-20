output "key_vault_id" {
  value = length(module.security) > 0 ? module.security[0].key_vault_id : null
}

output "key_vault_uri" {
  value = length(module.security) > 0 ? module.security[0].key_vault_uri : null
}

output "ai_foundry_key_vault_id" {
  value = length(module.security) > 0 ? module.security[0].ai_foundry_key_vault_id : null
}

output "ai_foundry_key_vault_uri" {
  value = length(module.security) > 0 ? module.security[0].ai_foundry_key_vault_uri : null
}

output "postgresql_id" {
  value = length(module.data) > 0 ? module.data[0].postgresql_id : null
}

output "storage_account_id" {
  value = length(module.data) > 0 ? module.data[0].storage_account_id : null
}

output "ai_foundry_storage_account_id" {
  value = length(module.data) > 0 ? module.data[0].ai_foundry_storage_account_id : null
}

output "acr_id" {
  value = length(module.data) > 0 ? module.data[0].acr_id : null
}

output "redis_id" {
  value = length(module.data) > 0 ? module.data[0].redis_id : null
}

output "managed_redis_id" {
  value = length(module.data) > 0 ? module.data[0].managed_redis_id : null
}

output "disk_encryption_set_id" {
  value = length(module.security) > 0 ? module.security[0].disk_encryption_set_id : null
}

output "azure_files_share_name" {
  value = length(module.data) > 0 ? module.data[0].azure_files_share_name : null
}
