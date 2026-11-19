output "bastion_vm_id" {
  value = google_compute_instance.bastion.id
}

output "bastion_private_ip" {
  value = google_compute_instance.bastion.network_interface[0].network_ip
}

output "bastion_ssh_private_key_pem" {
  value     = tls_private_key.bastion.private_key_pem
  sensitive = true
}