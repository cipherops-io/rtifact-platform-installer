data "azurerm_client_config" "current" {}

resource "azurerm_key_vault" "main" {
  name                       = "${substr(replace(var.name_prefix, "-", ""), 0, 18)}kv"
  location                   = var.location
  resource_group_name        = var.resource_group_name
  tenant_id                  = var.tenant_id != "" ? var.tenant_id : data.azurerm_client_config.current.tenant_id
  sku_name                   = var.sku
  soft_delete_retention_days = 7
  purge_protection_enabled   = true
  rbac_authorization_enabled = true
  tags                       = var.tags
}