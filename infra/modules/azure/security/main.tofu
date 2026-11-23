data "azurerm_client_config" "current" {}

resource "azurerm_key_vault" "main" {
  name                       = "${substr(replace(var.name_prefix, "-", ""), 0, 18)}kv"
  location                   = var.location
  resource_group_name        = var.resource_group_name
  tenant_id                  = var.tenant_id != "" ? var.tenant_id : data.azurerm_client_config.current.tenant_id
  sku_name                   = var.sku
  soft_delete_retention_days = var.key_vault.soft_delete_retention_days
  purge_protection_enabled   = var.key_vault.purge_protection_enabled
  rbac_authorization_enabled = var.key_vault.rbac_authorization_enabled
  tags                       = var.tags
}