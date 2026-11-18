resource "azurerm_container_registry" "main" {
  count = var.features.container_registry ? 1 : 0

  name                          = "${substr(replace(var.name_prefix, "-", ""), 0, 12)}${random_string.suffix[0].result}acr"
  resource_group_name           = var.resource_group_name
  location                      = var.location
  sku                           = var.acr_sku
  admin_enabled                 = false
  public_network_access_enabled = false
  tags                          = var.tags
}