locals {
  dns_resolver_enabled = var.features.private_dns_resolver && var.dns_resolver_inbound_subnet_cidr != null && var.dns_resolver_outbound_subnet_cidr != null
}

resource "azurerm_subnet" "dns_resolver_inbound" {
  count = local.dns_resolver_enabled ? 1 : 0

  name                 = "dns-resolver-inbound"
  resource_group_name  = var.resource_group_name
  virtual_network_name = azurerm_virtual_network.main.name
  address_prefixes     = [var.dns_resolver_inbound_subnet_cidr]

  delegation {
    name = "dns-resolver-delegation-inbound"
    service_delegation {
      name = "Microsoft.Network/dnsResolvers"
      actions = [
        "Microsoft.Network/virtualNetworks/subnets/join/action",
      ]
    }
  }
}

resource "azurerm_subnet" "dns_resolver_outbound" {
  count = local.dns_resolver_enabled ? 1 : 0

  name                 = "dns-resolver-outbound"
  resource_group_name  = var.resource_group_name
  virtual_network_name = azurerm_virtual_network.main.name
  address_prefixes     = [var.dns_resolver_outbound_subnet_cidr]

  delegation {
    name = "dns-resolver-delegation-outbound"
    service_delegation {
      name = "Microsoft.Network/dnsResolvers"
      actions = [
        "Microsoft.Network/virtualNetworks/subnets/join/action",
      ]
    }
  }
}

resource "azurerm_private_dns_resolver" "main" {
  count = local.dns_resolver_enabled ? 1 : 0

  name                = "${var.name_prefix}-dnsr"
  resource_group_name = var.resource_group_name
  location            = var.location
  virtual_network_id  = azurerm_virtual_network.main.id
  tags                = var.tags

  depends_on = [
    azurerm_subnet.node,
    azurerm_subnet.load_balancer,
    azurerm_subnet.private_endpoint,
    azurerm_subnet.management,
    azurerm_subnet.firewall,
    azurerm_subnet.gateway,
    azurerm_subnet.bastion,
    azurerm_subnet.dns_resolver_inbound,
    azurerm_subnet.dns_resolver_outbound,
  ]
}

resource "azurerm_private_dns_resolver_inbound_endpoint" "main" {
  count = local.dns_resolver_enabled && var.dns_resolver_inbound.enabled ? 1 : 0

  name                    = "${var.name_prefix}-dnsr-in"
  private_dns_resolver_id = azurerm_private_dns_resolver.main[0].id
  location                = var.location
  tags                    = var.tags

  ip_configurations {
    private_ip_allocation_method = var.dns_resolver_inbound.private_ip_allocation_method
    private_ip_address           = var.dns_resolver_inbound.private_ip_allocation_method == "Static" ? var.dns_resolver_inbound.private_ip_address : null
    subnet_id                    = azurerm_subnet.dns_resolver_inbound[0].id
  }
}

resource "azurerm_private_dns_resolver_outbound_endpoint" "main" {
  count = local.dns_resolver_enabled && var.dns_resolver_outbound_enabled ? 1 : 0

  name                    = "${var.name_prefix}-dnsr-out"
  private_dns_resolver_id = azurerm_private_dns_resolver.main[0].id
  location                = var.location
  tags                    = var.tags
  subnet_id               = azurerm_subnet.dns_resolver_outbound[0].id
}