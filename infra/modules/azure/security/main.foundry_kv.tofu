# Legacy Hub (azurerm_ai_foundry) provisions Hub MSI access policies during create; RBAC-only vaults reject that.
resource "azurerm_key_vault" "foundry" {
  count = var.features.ai_foundry ? 1 : 0

  name                       = "${substr(replace(var.name_prefix, "-", ""), 0, 16)}fkv"
  location                   = var.location
  resource_group_name        = var.resource_group_name
  tenant_id                  = var.tenant_id != "" ? var.tenant_id : data.azurerm_client_config.current.tenant_id
  sku_name                   = var.foundry_key_vault.sku
  soft_delete_retention_days = var.foundry_key_vault.soft_delete_retention_days
  purge_protection_enabled   = var.foundry_key_vault.purge_protection_enabled
  rbac_authorization_enabled = false
  tags                       = var.tags
}

# ARM control plane: Hub create calls accessPolicies/add; data-plane access policy alone is insufficient.
resource "azurerm_role_assignment" "foundry_deployer_contributor" {
  count = var.features.ai_foundry && var.foundry_key_vault.grant_deployer_contributor ? 1 : 0

  scope                = azurerm_key_vault.foundry[0].id
  role_definition_name = "Key Vault Contributor"
  principal_id         = data.azurerm_client_config.current.object_id
}

resource "azurerm_key_vault_access_policy" "foundry_deployer" {
  count = var.features.ai_foundry ? 1 : 0

  key_vault_id = azurerm_key_vault.foundry[0].id
  tenant_id    = data.azurerm_client_config.current.tenant_id
  object_id    = data.azurerm_client_config.current.object_id

  key_permissions         = var.foundry_key_vault.deployer_key_permissions
  secret_permissions      = var.foundry_key_vault.deployer_secret_permissions
  certificate_permissions = var.foundry_key_vault.deployer_certificate_permissions
}