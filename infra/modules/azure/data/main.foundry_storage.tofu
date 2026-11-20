resource "random_string" "foundry_suffix" {
  count = var.features.ai_foundry ? 1 : 0

  length  = 4
  special = false
  upper   = false
}

# Azure AI Foundry cannot use the ADLS (HNS) corpus account; this account is non-HNS StorageV2.
resource "azurerm_storage_account" "foundry" {
  count = var.features.ai_foundry ? 1 : 0

  name                     = "${substr(replace(var.name_prefix, "-", ""), 0, 12)}ml${random_string.foundry_suffix[0].result}"
  resource_group_name      = var.resource_group_name
  location                 = var.location
  account_tier             = var.foundry_storage.account_tier
  account_replication_type = var.foundry_storage.account_replication_type
  account_kind             = "StorageV2"
  min_tls_version          = "TLS1_2"
  public_network_access    = var.foundry_storage.public_network_access
  is_hns_enabled           = false
  tags                     = var.tags
}

resource "azurerm_private_endpoint" "foundry_blob" {
  count = var.features.ai_foundry && local.foundry_blob_pe ? 1 : 0

  name                = "${var.name_prefix}-foundry-blob-pe"
  location            = var.location
  resource_group_name = var.resource_group_name
  subnet_id           = var.private_endpoint_subnet_id
  tags                = var.tags

  private_service_connection {
    name                           = "foundry-blob"
    private_connection_resource_id = azurerm_storage_account.foundry[0].id
    subresource_names              = ["blob"]
    is_manual_connection           = false
  }

  dynamic "private_dns_zone_group" {
    for_each = local.blob_dns ? [1] : []
    content {
      name                 = "default"
      private_dns_zone_ids = [azurerm_private_dns_zone.blob[0].id]
    }
  }
}