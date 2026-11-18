resource "azurerm_storage_account" "main" {
  count = var.features.storage_account ? 1 : 0

  name                     = "${substr(replace(var.name_prefix, "-", ""), 0, 12)}${random_string.suffix[0].result}"
  resource_group_name      = var.resource_group_name
  location                 = var.location
  account_tier             = "Premium"
  account_replication_type = "LRS"
  account_kind             = "BlockBlobStorage"
  min_tls_version          = "TLS1_2"
  public_network_access    = "Disabled"
  is_hns_enabled           = true
  tags                     = var.tags

  blob_properties {
    delete_retention_policy {
      days = 30
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