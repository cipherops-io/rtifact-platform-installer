resource "azurerm_storage_account" "main" {
  count = var.features.storage_account ? 1 : 0

  name                     = "${substr(replace(var.name_prefix, "-", ""), 0, 12)}${random_string.suffix[0].result}"
  resource_group_name      = var.resource_group_name
  location                 = var.location
  account_tier             = var.storage.account_tier
  account_replication_type = var.storage.account_replication_type
  account_kind             = var.storage.account_kind
  min_tls_version          = "TLS1_2"
  public_network_access    = var.storage.public_network_access
  is_hns_enabled           = var.storage.is_hns_enabled
  tags                     = var.tags

  blob_properties {
    delete_retention_policy {
      days = var.storage.delete_retention_days
    }
  }
}

resource "azurerm_storage_container" "corpus" {
  count = var.features.storage_account ? 1 : 0

  name                  = "corpus"
  storage_account_id    = azurerm_storage_account.main[0].id
  container_access_type = "private"
}

resource "azurerm_storage_container" "backup" {
  count = var.features.storage_account ? 1 : 0

  name                  = "backup"
  storage_account_id    = azurerm_storage_account.main[0].id
  container_access_type = "private"
}