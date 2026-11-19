data "azurerm_client_config" "des" {}

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
  key_type     = "RSA"
  key_size     = 3072

  key_opts = ["decrypt", "encrypt", "sign", "unwrapKey", "verify", "wrapKey"]
}

resource "azurerm_disk_encryption_set" "main" {
  count = var.features.disk_encryption_set ? 1 : 0

  name                = "${var.name_prefix}-des"
  resource_group_name = var.resource_group_name
  location            = var.location
  key_vault_key_id    = azurerm_key_vault_key.des[0].id
  tags                = var.tags

  identity {
    type = "UserAssigned"
    identity_ids = [
      azurerm_user_assigned_identity.des[0].id,
    ]
  }
}

resource "azurerm_key_vault_access_policy" "des" {
  count = var.features.disk_encryption_set ? 1 : 0

  key_vault_id = azurerm_key_vault.main.id
  tenant_id    = data.azurerm_client_config.des.tenant_id
  object_id    = azurerm_user_assigned_identity.des[0].principal_id

  key_permissions = [
    "Get",
    "WrapKey",
    "UnwrapKey",
  ]
}