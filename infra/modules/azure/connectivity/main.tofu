locals {
  create_vpn = var.vpn_enabled && !var.express_route_enabled
  create_er  = var.express_route_enabled && !var.vpn_enabled
}

resource "azurerm_public_ip" "vpn" {
  count = local.create_vpn ? 1 : 0

  name                = "${var.name_prefix}-vpngw-pip"
  location            = var.location
  resource_group_name = var.resource_group_name
  allocation_method   = "Static"
  sku                 = "Standard"
  tags                = var.tags
}

resource "azurerm_virtual_network_gateway" "vpn" {
  count = local.create_vpn ? 1 : 0

  name                = "${var.name_prefix}-vpngw"
  location            = var.location
  resource_group_name = var.resource_group_name
  type                = var.vpn.gateway_type
  vpn_type            = var.vpn.vpn_type
  active_active       = var.vpn.active_active
  sku                 = var.vpn.sku
  tags                = var.tags

  ip_configuration {
    name                          = "default"
    public_ip_address_id          = azurerm_public_ip.vpn[0].id
    private_ip_address_allocation = "Dynamic"
    subnet_id                     = var.gateway_subnet_id
  }
}

resource "azurerm_public_ip" "er" {
  count = local.create_er ? 1 : 0

  name                = "${var.name_prefix}-ergw-pip"
  location            = var.location
  resource_group_name = var.resource_group_name
  allocation_method   = "Static"
  sku                 = "Standard"
  tags                = var.tags
}

resource "azurerm_virtual_network_gateway" "express_route" {
  count = local.create_er ? 1 : 0

  name                = "${var.name_prefix}-ergw"
  location            = var.location
  resource_group_name = var.resource_group_name
  type                = "ExpressRoute"
  sku                 = var.express_route.sku
  tags                = var.tags

  ip_configuration {
    name                          = "default"
    public_ip_address_id          = azurerm_public_ip.er[0].id
    private_ip_address_allocation = "Dynamic"
    subnet_id                     = var.gateway_subnet_id
  }
}

resource "azurerm_virtual_network_gateway_connection" "er" {
  count = local.create_er && var.express_route.express_route_circuit_id != null && var.express_route.express_route_circuit_id != "" ? 1 : 0

  name                       = "${var.name_prefix}-er-conn"
  location                   = var.location
  resource_group_name        = var.resource_group_name
  type                       = "ExpressRoute"
  virtual_network_gateway_id = azurerm_virtual_network_gateway.express_route[0].id
  express_route_circuit_id   = var.express_route.express_route_circuit_id
  routing_weight             = 0
  tags                       = var.tags
}