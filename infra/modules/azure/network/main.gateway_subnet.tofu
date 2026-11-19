locals {
  gateway_enabled = (var.features.azure_vpn_gateway || var.features.azure_express_route_gateway) && var.gateway_subnet_cidr != null
}

resource "azurerm_subnet" "gateway" {
  count = local.gateway_enabled ? 1 : 0

  name                 = "GatewaySubnet"
  resource_group_name  = var.resource_group_name
  virtual_network_name = azurerm_virtual_network.main.name
  address_prefixes     = [var.gateway_subnet_cidr]
}