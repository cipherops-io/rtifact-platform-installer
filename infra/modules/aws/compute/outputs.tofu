output "bastion_vm_id" {
  value = aws_instance.bastion.id
}

output "bastion_private_ip" {
  value = aws_instance.bastion.private_ip
}

output "bastion_ssh_private_key_pem" {
  value     = tls_private_key.bastion.private_key_pem
  sensitive = true
}