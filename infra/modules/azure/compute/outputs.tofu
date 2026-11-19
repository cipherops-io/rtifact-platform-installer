output "bastion_vm_id" {
  value = azurerm_linux_virtual_machine.bastion.id
}

output "bastion_private_ip" {
  value = azurerm_network_interface.bastion.private_ip_address
}

output "bastion_ssh_private_key_pem" {
  value     = tls_private_key.bastion.private_key_pem
  sensitive = true
}