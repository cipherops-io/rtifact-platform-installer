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
