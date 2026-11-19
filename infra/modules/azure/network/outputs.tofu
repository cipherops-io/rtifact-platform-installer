output "vnet_id" {
  value = azurerm_virtual_network.main.id
}

output "vnet_name" {
  value = azurerm_virtual_network.main.name
}

output "node_subnet_id" {
  value = azurerm_subnet.node.id
}

output "load_balancer_subnet_id" {
  value = azurerm_subnet.load_balancer.id
}

output "private_endpoint_subnet_id" {
  value = azurerm_subnet.private_endpoint.id
}

output "management_subnet_id" {
  value = azurerm_subnet.management.id
}

output "nat_gateway_id" {
  value = azurerm_nat_gateway.main.id
}

output "firewall_subnet_id" {
  value = length(azurerm_subnet.firewall) > 0 ? azurerm_subnet.firewall[0].id : null
}

output "gateway_subnet_id" {
  value = length(azurerm_subnet.gateway) > 0 ? azurerm_subnet.gateway[0].id : null
}

output "bastion_subnet_id" {
  value = length(azurerm_subnet.bastion) > 0 ? azurerm_subnet.bastion[0].id : null
}

output "private_dns_resolver_id" {
  value = length(azurerm_private_dns_resolver.main) > 0 ? azurerm_private_dns_resolver.main[0].id : null
}