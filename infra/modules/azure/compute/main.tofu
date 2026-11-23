locals {
  public_ip_enabled = var.access == "public_ip"
}

resource "azurerm_public_ip" "bastion" {
  count = local.public_ip_enabled ? 1 : 0

  name                = "${var.name_prefix}-bastion-vm-pip"
  location            = var.location
  resource_group_name = var.resource_group_name
  allocation_method   = "Static"
  sku                 = "Standard"
  tags                = var.tags
}

resource "azurerm_network_interface" "bastion" {
  name                = "${var.name_prefix}-bastion-nic"
  location            = var.location
  resource_group_name = var.resource_group_name
  tags                = var.tags

  ip_configuration {
    name                          = "internal"
    subnet_id                     = var.subnet_id
    private_ip_address_allocation = var.private_ip.allocation_method
    private_ip_address            = var.private_ip.allocation_method == "Static" ? var.private_ip.address : null
    public_ip_address_id          = local.public_ip_enabled ? azurerm_public_ip.bastion[0].id : null
  }
}

resource "azurerm_network_security_group" "bastion" {
  count = local.public_ip_enabled ? 1 : 0

  name                = "${var.name_prefix}-bastion-nsg"
  location            = var.location
  resource_group_name = var.resource_group_name
  tags                = var.tags

  security_rule {
    name                       = "allow-ssh"
    priority                   = 100
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "22"
    source_address_prefixes    = var.allowed_ssh_cidrs
    destination_address_prefix = "*"
  }
}

resource "azurerm_network_interface_security_group_association" "bastion" {
  count = local.public_ip_enabled ? 1 : 0

  network_interface_id      = azurerm_network_interface.bastion.id
  network_security_group_id = azurerm_network_security_group.bastion[0].id
}

resource "azurerm_linux_virtual_machine" "bastion" {
  name                = "${var.name_prefix}-bastion"
  location            = var.location
  resource_group_name = var.resource_group_name
  size                = var.vm_size
  admin_username      = var.admin_username
  tags                = var.tags

  network_interface_ids = [
    azurerm_network_interface.bastion.id,
  ]

  admin_ssh_key {
    username   = var.admin_username
    public_key = tls_private_key.bastion.public_key_openssh
  }

  os_disk {
    name                 = "${var.name_prefix}-bastion-os"
    caching              = var.os_disk_caching
    storage_account_type = var.os_disk_type
    disk_size_gb         = var.os_disk_size_gb
  }

  source_image_reference {
    publisher = var.image.publisher
    offer     = var.image.offer
    sku       = var.image.sku
    version   = var.image.version
  }

  disable_password_authentication = true
}

resource "tls_private_key" "bastion" {
  algorithm = var.ssh_key.algorithm
  rsa_bits  = var.ssh_key.algorithm == "RSA" ? var.ssh_key.rsa_bits : null
}
