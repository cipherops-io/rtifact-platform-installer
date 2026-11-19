output "vpn_gateway_id" {
  value = length(azurerm_virtual_network_gateway.vpn) > 0 ? azurerm_virtual_network_gateway.vpn[0].id : null
}

output "express_route_gateway_id" {
  value = length(azurerm_virtual_network_gateway.express_route) > 0 ? azurerm_virtual_network_gateway.express_route[0].id : null
}