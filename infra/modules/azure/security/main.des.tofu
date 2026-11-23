resource "azurerm_user_assigned_identity" "des" {
  count = var.features.disk_encryption_set ? 1 : 0

  name                = "${var.name_prefix}-des-id"
  resource_group_name = var.resource_group_name
  location            = var.location
  tags                = var.tags
}

resource "azurerm_key_vault_key" "des" {
  count = var.features.disk_encryption_set ? 1 : 0

  name         = "${replace(var.name_prefix, "-", "")}-des"
  key_vault_id = azurerm_key_vault.main.id
  key_type     = var.des_key.key_type
  key_size     = var.des_key.key_size

  key_opts = ["decrypt", "encrypt", "sign", "unwrapKey", "verify", "wrapKey"]

  depends_on = [
    azurerm_role_assignment.deployer_crypto_officer,
    azurerm_role_assignment.des_crypto_user,
  ]
}

resource "azurerm_disk_encryption_set" "main" {
  count = var.features.disk_encryption_set ? 1 : 0

  name                      = "${var.name_prefix}-des"
  resource_group_name       = var.resource_group_name
  location                  = var.location
  key_vault_key_id          = var.des_key.auto_key_rotation_enabled ? azurerm_key_vault_key.des[0].versionless_id : azurerm_key_vault_key.des[0].id
  auto_key_rotation_enabled = var.des_key.auto_key_rotation_enabled
  tags                      = var.tags

  identity {
    type = "UserAssigned"
    identity_ids = [
      azurerm_user_assigned_identity.des[0].id,
    ]
  }

  depends_on = [
    azurerm_role_assignment.des_crypto_user,
    azurerm_key_vault_key.des,
  ]
}