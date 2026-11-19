resource "azurerm_ai_foundry" "main" {
  name                  = "${var.name_prefix}-foundry"
  location              = var.location
  resource_group_name   = var.resource_group_name
  storage_account_id    = var.storage_account_id
  key_vault_id          = var.key_vault_id
  public_network_access = var.public_network_access

  identity {
    type = "SystemAssigned"
  }

  container_registry_id = var.container_registry_id

  tags = var.tags
}