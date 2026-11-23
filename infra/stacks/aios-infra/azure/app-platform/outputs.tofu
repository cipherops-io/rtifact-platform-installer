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

output "bastion_public_ip" {
  value = length(module.compute) > 0 ? module.compute[0].bastion_public_ip : null
}

output "bastion_admin_username" {
  value = length(module.compute) > 0 ? module.compute[0].bastion_admin_username : null
}

output "bastion_vm_id" {
  value = length(module.compute) > 0 ? module.compute[0].bastion_vm_id : null
}

output "bastion_access" {
  value = length(module.compute) > 0 ? var.bastion.access : null
}

output "bastion_ssh_command" {
  description = "SSH into the bastion; the key file is written next to the kubeconfig by `rtifact outputs`."
  value       = local.bastion_ssh_command
}
