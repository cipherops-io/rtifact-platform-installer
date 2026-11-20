output "key_vault_id" {
  value = azurerm_key_vault.main.id
}

output "key_vault_uri" {
  value = azurerm_key_vault.main.vault_uri
}

output "disk_encryption_set_id" {
  value = length(azurerm_disk_encryption_set.main) > 0 ? azurerm_disk_encryption_set.main[0].id : null
}

output "ai_foundry_key_vault_id" {
  value = try(azurerm_key_vault.foundry[0].id, null)
}

output "ai_foundry_key_vault_uri" {
  value = try(azurerm_key_vault.foundry[0].vault_uri, null)
}