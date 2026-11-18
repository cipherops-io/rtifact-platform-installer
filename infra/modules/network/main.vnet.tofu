resource "azurerm_virtual_network" "main" {
  name                = "${var.name_prefix}-vnet"
  location            = var.location
  resource_group_name = var.resource_group_name
  address_space       = [var.vnet_cidr]
  tags                = var.tags
}

# Azure CNI Overlay (see modules/aks) does not use a delegated node subnet; delegation causes SubnetIsDelegated on create.
resource "azurerm_subnet" "node" {
  name                 = "aks-nodes"
  resource_group_name  = var.resource_group_name
  virtual_network_name = azurerm_virtual_network.main.name
  address_prefixes     = [var.node_subnet_cidr]
}

resource "azurerm_subnet" "load_balancer" {
  name                 = "load-balancer"
  resource_group_name  = var.resource_group_name
  virtual_network_name = azurerm_virtual_network.main.name
  address_prefixes     = [var.load_balancer_subnet_cidr]
}

resource "azurerm_subnet" "private_endpoint" {
  name                 = "private-endpoints"
  resource_group_name  = var.resource_group_name
  virtual_network_name = azurerm_virtual_network.main.name
  address_prefixes     = [var.private_endpoint_subnet_cidr]
}

resource "azurerm_subnet" "management" {
  name                 = "management"
  resource_group_name  = var.resource_group_name
  virtual_network_name = azurerm_virtual_network.main.name
  address_prefixes     = [var.management_subnet_cidr]
}