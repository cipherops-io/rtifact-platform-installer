output "resource_group_name" {
  value = var.features.resource_group ? module.naming[0].resource_group_name : null
}

output "resource_group_id" {
  value = var.features.resource_group ? module.naming[0].resource_group_id : null
}

output "tags" {
  value = var.features.resource_group ? module.naming[0].tags : {}
}

output "vnet_id" {
  value = var.features.virtual_network ? module.network[0].vnet_id : null
}

output "node_subnet_id" {
  value = var.features.virtual_network ? module.network[0].node_subnet_id : null
}

output "load_balancer_subnet_id" {
  value = var.features.virtual_network ? module.network[0].load_balancer_subnet_id : null
}

output "private_endpoint_subnet_id" {
  value = var.features.virtual_network ? module.network[0].private_endpoint_subnet_id : null
}

output "management_subnet_id" {
  value = var.features.virtual_network ? module.network[0].management_subnet_id : null
}

output "firewall_id" {
  value = length(module.firewall) > 0 ? module.firewall[0].firewall_id : null
}

output "vpn_gateway_id" {
  value = length(module.connectivity) > 0 ? module.connectivity[0].vpn_gateway_id : null
}

output "express_route_gateway_id" {
  value = length(module.connectivity) > 0 ? module.connectivity[0].express_route_gateway_id : null
}

output "bastion_host_id" {
  value = length(module.bastion_host) > 0 ? module.bastion_host[0].bastion_host_id : null
}

output "azure_bastion_name" {
  value = length(module.bastion_host) > 0 ? module.bastion_host[0].bastion_host_name : null
}

output "private_dns_resolver_id" {
  value = var.features.virtual_network ? module.network[0].private_dns_resolver_id : null
}
