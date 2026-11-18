locals {
  key_vault_pe_enabled = var.features.key_vault_private_endpoint && var.features.key_vault
}

resource "azurerm_private_dns_zone" "vault" {
  count = local.key_vault_pe_enabled ? 1 : 0

  name                = "privatelink.vaultcore.azure.net"
  resource_group_name = module.naming[0].resource_group_name
  tags                = module.naming[0].tags
}

resource "azurerm_private_dns_zone_virtual_network_link" "vault" {
  count = local.key_vault_pe_enabled ? 1 : 0

  name                = "${var.name_prefix}-kv-dns"
  private_dns_zone_id = azurerm_private_dns_zone.vault[0].id
  virtual_network_id  = module.network[0].vnet_id
}

module "key_vault_pe" {
  count = local.key_vault_pe_enabled ? 1 : 0

  source                         = "../../modules/private-endpoint"
  name                           = "${var.name_prefix}-kv-pe"
  location                       = var.location
  resource_group_name            = module.naming[0].resource_group_name
  subnet_id                      = module.network[0].private_endpoint_subnet_id
  private_connection_resource_id = module.security[0].key_vault_id
  subresource_names              = ["vault"]
  private_dns_zone_ids           = [azurerm_private_dns_zone.vault[0].id]
  tags                           = module.naming[0].tags

  depends_on = [module.security]
}