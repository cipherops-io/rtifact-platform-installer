# RBAC for RBAC-enabled vaults: deployer needs data-plane roles before key create (no vault↔key cycle).
# Assignment scope = vault.id (depends on vault only); keys depend_on assignments, not the reverse.

resource "azurerm_role_assignment" "deployer_crypto_officer" {
  count = var.features.disk_encryption_set ? 1 : 0

  scope                = azurerm_key_vault.main.id
  role_definition_name = "Key Vault Crypto Officer"
  principal_id         = data.azurerm_client_config.current.object_id
}

resource "azurerm_role_assignment" "des_crypto_user" {
  count = var.features.disk_encryption_set ? 1 : 0

  scope                = azurerm_key_vault.main.id
  role_definition_name = "Key Vault Crypto Service Encryption User"
  principal_id         = azurerm_user_assigned_identity.des[0].principal_id
}