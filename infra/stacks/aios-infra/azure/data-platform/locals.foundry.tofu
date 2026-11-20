locals {
  foundry_key_vault_settings = {
    sku                              = coalesce(try(var.key_vault.foundry.sku, null), var.key_vault.sku)
    soft_delete_retention_days       = var.key_vault.foundry.soft_delete_retention_days
    purge_protection_enabled         = var.key_vault.foundry.purge_protection_enabled
    grant_deployer_contributor       = var.key_vault.foundry.grant_deployer_contributor
    deployer_key_permissions         = var.key_vault.foundry.deployer_key_permissions
    deployer_secret_permissions      = var.key_vault.foundry.deployer_secret_permissions
    deployer_certificate_permissions = var.key_vault.foundry.deployer_certificate_permissions
  }

  foundry_storage_settings = {
    account_tier             = var.ai_foundry.storage.account_tier
    account_replication_type = var.ai_foundry.storage.account_replication_type
    public_network_access    = var.ai_foundry.storage.public_network_access
  }
}