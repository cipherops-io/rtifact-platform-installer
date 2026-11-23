locals {
  bastion_subnet_enabled = var.bastion_managed && var.azure_bastion_subnet_cidr != null
}

resource "azurerm_subnet" "bastion" {
  count = local.bastion_subnet_enabled ? 1 : 0

  name                 = "AzureBastionSubnet"
  resource_group_name  = var.resource_group_name
  virtual_network_name = azurerm_virtual_network.main.name
  address_prefixes     = [var.azure_bastion_subnet_cidr]

  lifecycle {
    precondition {
      condition     = tonumber(split("/", var.azure_bastion_subnet_cidr)[1]) <= 26
      error_message = "network.azure_bastion_subnet_cidr must be /26 or larger for Azure Bastion."
    }
  }
}