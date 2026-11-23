locals {
  acr_private_endpoint      = var.features.container_registry && var.acr_sku == "Premium"
  acr_public_network_access = coalesce(var.acr.public_network_access_enabled, var.acr_sku != "Premium")
}

resource "azurerm_container_registry" "main" {
  count = var.features.container_registry ? 1 : 0

  name                          = "${substr(replace(var.name_prefix, "-", ""), 0, 12)}${random_string.suffix[0].result}acr"
  resource_group_name           = var.resource_group_name
  location                      = var.location
  sku                           = var.acr_sku
  admin_enabled                 = var.acr.admin_enabled
  public_network_access_enabled = local.acr_public_network_access
  tags                          = var.tags
}