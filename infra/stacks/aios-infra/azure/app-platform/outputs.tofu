output "cluster_name" {
  value = length(module.aks) > 0 ? module.aks[0].cluster_name : null
}

output "aks_name" {
  value = length(module.aks) > 0 ? module.aks[0].cluster_name : null
}

output "oidc_issuer_url" {
  value = length(module.aks) > 0 ? module.aks[0].oidc_issuer_url : null
}

output "kubeconfig_raw" {
  value     = length(module.aks) > 0 ? module.aks[0].kubeconfig_raw : null
  sensitive = true
}

output "kube_admin_config_raw" {
  value     = length(module.aks) > 0 ? module.aks[0].kube_admin_config_raw : null
  sensitive = true
}

output "bastion_private_ip" {
  value = length(module.compute) > 0 ? module.compute[0].bastion_private_ip : null
}

output "bastion_ssh_private_key_pem" {
  value     = length(module.compute) > 0 ? module.compute[0].bastion_ssh_private_key_pem : null
  sensitive = true
}
