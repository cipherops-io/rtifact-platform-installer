locals {
  firewall_enabled = var.features.azure_firewall && var.azure_firewall_subnet_cidr != null
}

resource "azurerm_subnet" "firewall" {
  count = local.firewall_enabled ? 1 : 0

  name                 = "AzureFirewallSubnet"
  resource_group_name  = var.resource_group_name
  virtual_network_name = azurerm_virtual_network.main.name
  address_prefixes     = [var.azure_firewall_subnet_cidr]
}