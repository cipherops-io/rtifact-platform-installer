output "key_vault_id" {
  value = length(module.security) > 0 ? module.security[0].key_vault_id : null
}

output "key_vault_uri" {
  value = length(module.security) > 0 ? module.security[0].key_vault_uri : null
}

output "postgresql_id" {
  value = length(module.data) > 0 ? module.data[0].postgresql_id : null
}

output "storage_account_id" {
  value = length(module.data) > 0 ? module.data[0].storage_account_id : null
}

output "acr_id" {
  value = length(module.data) > 0 ? module.data[0].acr_id : null
}
