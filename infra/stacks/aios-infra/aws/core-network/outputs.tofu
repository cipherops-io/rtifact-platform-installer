output "tags" {
  value = local.net_tags
}

output "vpc_id" {
  value = length(module.network) > 0 ? module.network[0].vpc_id : null
}

output "vpc_cidrs" {
  value = length(module.network) > 0 ? module.network[0].vpc_cidrs : []
}

output "subnet_ids_by_tier" {
  value = length(module.network) > 0 ? module.network[0].subnet_ids_by_tier : {}
}

output "availability_zones" {
  value = length(module.network) > 0 ? module.network[0].availability_zones : []
}

output "route_table_ids" {
  value = length(module.network) > 0 ? module.network[0].route_table_ids : null
}

output "nat_gateway_public_ips" {
  value = length(module.network) > 0 ? module.network[0].nat_gateway_public_ips : []
}

output "bastion_instance_id" {
  value = length(module.bastion) > 0 ? module.bastion[0].instance_id : null
}

output "bastion_private_ip" {
  value = length(module.bastion) > 0 ? module.bastion[0].private_ip : null
}

output "bastion_public_ip" {
  value = length(module.bastion) > 0 ? module.bastion[0].public_ip : null
}

output "bastion_ssh_private_key_pem" {
  description = "Generated bastion SSH private key; rtifact outputs writes it to <name_prefix>-bastion.pem."
  value       = length(module.bastion) > 0 ? module.bastion[0].ssh_private_key_pem : null
  sensitive   = true
}

output "bastion_ssh_command" {
  description = "Command to reach the bastion (ssh for access=ssh, SSM Session Manager for access=ssm)."
  value       = local.bastion_ssh_command
}
