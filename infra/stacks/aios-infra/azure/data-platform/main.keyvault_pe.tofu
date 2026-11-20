locals {
  key_vault_pe_enabled = var.features.key_vault_private_endpoint && var.features.key_vault
}

resource "azurerm_private_dns_zone" "vault" {
  count = local.key_vault_pe_enabled ? 1 : 0

  name                = "privatelink.vaultcore.azure.net"
  resource_group_name = local.core.resource_group_name
  tags                = local.core.tags
}

resource "azurerm_private_dns_zone_virtual_network_link" "vault" {
  count = local.key_vault_pe_enabled ? 1 : 0

  name                = "${var.name_prefix}-kv-dns"
  private_dns_zone_id = azurerm_private_dns_zone.vault[0].id
  virtual_network_id  = local.core.vnet_id
}

module "key_vault_pe" {
  count = local.key_vault_pe_enabled ? 1 : 0

  source                         = "../../../../modules/azure/private-endpoint"
  name                           = "${var.name_prefix}-kv-pe"
  location                       = var.location
  resource_group_name            = local.core.resource_group_name
  subnet_id                      = local.core.private_endpoint_subnet_id
  private_connection_resource_id = module.security[0].key_vault_id
  subresource_names              = ["vault"]
  private_dns_zone_ids           = [azurerm_private_dns_zone.vault[0].id]
  tags                           = local.core.tags

  depends_on = [module.security]
}

module "foundry_key_vault_pe" {
  count = local.key_vault_pe_enabled && var.features.ai_foundry ? 1 : 0

  source                         = "../../../../modules/azure/private-endpoint"
  name                           = "${var.name_prefix}-foundry-kv-pe"
  location                       = var.location
  resource_group_name            = local.core.resource_group_name
  subnet_id                      = local.core.private_endpoint_subnet_id
  private_connection_resource_id = module.security[0].ai_foundry_key_vault_id
  subresource_names              = ["vault"]
  private_dns_zone_ids           = [azurerm_private_dns_zone.vault[0].id]
  tags                           = local.core.tags

  depends_on = [module.security]
}
