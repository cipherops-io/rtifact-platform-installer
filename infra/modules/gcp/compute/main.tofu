resource "tls_private_key" "bastion" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "google_compute_instance" "bastion" {
  name         = "${var.name_prefix}-bastion"
  machine_type = var.vm_size
  zone         = "${var.location}-a"

  boot_disk {
    initialize_params {
      image = "ubuntu-os-cloud/ubuntu-2204-lts"
      size  = var.os_disk_size_gb
    }
  }

  network_interface {
    subnetwork = var.subnet_id
  }

  metadata = {
    ssh-keys = "ubuntu:${tls_private_key.bastion.public_key_openssh}"
  }

  tags = ["bastion"]
}