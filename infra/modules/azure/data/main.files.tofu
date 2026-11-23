resource "random_string" "files_suffix" {
  count = var.features.azure_files_share ? 1 : 0

  length  = 4
  special = false
  upper   = false
}

resource "azurerm_storage_account" "files" {
  count = var.features.azure_files_share ? 1 : 0

  name                     = "${substr(replace(var.name_prefix, "-", ""), 0, 12)}f${random_string.files_suffix[0].result}"
  resource_group_name      = var.resource_group_name
  location                 = var.location
  account_tier             = "Premium"
  account_replication_type = var.azure_files_zone_redundant ? "ZRS" : "LRS"
  account_kind             = "FileStorage"
  min_tls_version          = "TLS1_2"
  public_network_access    = var.azure_files.public_network_access
  tags                     = var.tags
}

resource "azurerm_storage_share" "aios" {
  count = var.features.azure_files_share ? 1 : 0

  name               = var.azure_files_share_name
  storage_account_id = azurerm_storage_account.files[0].id
  quota              = var.azure_files_quota_gb
  enabled_protocol   = var.azure_files.enabled_protocol
}

resource "azurerm_private_dns_zone" "file" {
  count = local.file_dns ? 1 : 0

  name                = "privatelink.file.core.windows.net"
  resource_group_name = var.dns_resource_group_name
  tags                = var.tags
}

resource "azurerm_private_dns_zone_virtual_network_link" "file" {
  count = local.file_dns ? 1 : 0

  name                 = "${var.name_prefix}-file-dns"
  private_dns_zone_id  = azurerm_private_dns_zone.file[0].id
  virtual_network_id   = var.vnet_id
  registration_enabled = false
}

resource "azurerm_private_endpoint" "files" {
  count = var.features.azure_files_share ? 1 : 0

  name                = "${var.name_prefix}-files-pe"
  location            = var.location
  resource_group_name = var.resource_group_name
  subnet_id           = var.private_endpoint_subnet_id
  tags                = var.tags

  private_service_connection {
    name                           = "files"
    private_connection_resource_id = azurerm_storage_account.files[0].id
    subresource_names              = ["file"]
    is_manual_connection           = var.private_endpoint_connection.is_manual_connection
    request_message                = var.private_endpoint_connection.request_message
  }

  dynamic "private_dns_zone_group" {
    for_each = local.file_dns ? [1] : []
    content {
      name                 = "default"
      private_dns_zone_ids = [azurerm_private_dns_zone.file[0].id]
    }
  }
}