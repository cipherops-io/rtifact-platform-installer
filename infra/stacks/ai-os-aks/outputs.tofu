output "resource_group_name" {
  value = length(module.naming) > 0 ? module.naming[0].resource_group_name : null
}

output "aks_name" {
  value = length(module.aks) > 0 ? module.aks[0].cluster_name : null
}

output "key_vault_uri" {
  value = length(module.security) > 0 ? module.security[0].key_vault_uri : null
}

output "bastion_private_ip" {
  value = length(module.compute) > 0 ? module.compute[0].bastion_private_ip : null
}