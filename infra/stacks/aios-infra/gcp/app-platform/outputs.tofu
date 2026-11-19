output "cluster_name" {
  value = length(module.gke) > 0 ? module.gke[0].cluster_name : null
}

output "aks_name" {
  value = length(module.gke) > 0 ? module.gke[0].cluster_name : null
}

output "oidc_issuer_url" {
  value = length(module.gke) > 0 ? module.gke[0].oidc_issuer_url : null
}

output "bastion_private_ip" {
  value = length(module.compute) > 0 ? module.compute[0].bastion_private_ip : null
}

output "bastion_ssh_private_key_pem" {
  value     = length(module.compute) > 0 ? module.compute[0].bastion_ssh_private_key_pem : null
  sensitive = true
}
