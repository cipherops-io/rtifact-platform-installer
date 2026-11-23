output "bastion_vm_id" {
  value = azurerm_linux_virtual_machine.bastion.id
}

output "bastion_private_ip" {
  value = azurerm_network_interface.bastion.private_ip_address
}

output "bastion_public_ip" {
  value = local.public_ip_enabled ? azurerm_public_ip.bastion[0].ip_address : null
}

output "bastion_admin_username" {
  value = var.admin_username
}

output "bastion_ssh_private_key_pem" {
  value     = var.ssh_key.algorithm == "RSA" ? tls_private_key.bastion.private_key_pem : tls_private_key.bastion.private_key_openssh
  sensitive = true
}
