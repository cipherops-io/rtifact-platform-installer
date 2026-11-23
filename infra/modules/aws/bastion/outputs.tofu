output "instance_id" {
  value = aws_instance.bastion.id
}

output "private_ip" {
  value = aws_instance.bastion.private_ip
}

output "public_ip" {
  value = var.bastion.public_ip ? aws_eip.bastion[0].public_ip : null
}

output "access" {
  value = var.bastion.access
}

output "admin_username" {
  value = var.bastion.admin_username
}

output "ssh_port" {
  value = var.bastion.ssh_port
}

output "ssh_private_key_pem" {
  description = "Generated SSH private key (PEM for RSA, OpenSSH for ED25519); null when access is ssm or a public key was supplied."
  value       = local.generate_key ? (var.bastion.ssh_key.algorithm == "RSA" ? tls_private_key.bastion[0].private_key_pem : tls_private_key.bastion[0].private_key_openssh) : null
  sensitive   = true
}
