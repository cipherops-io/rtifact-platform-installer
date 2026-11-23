mock_provider "azurerm" {
  mock_resource "azurerm_network_interface" {
    defaults = {
      id = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.Network/networkInterfaces/t-bastion-nic"
    }
  }
  mock_resource "azurerm_public_ip" {
    defaults = {
      id = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.Network/publicIPAddresses/t-bastion-vm-pip"
    }
  }
  mock_resource "azurerm_network_security_group" {
    defaults = {
      id = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.Network/networkSecurityGroups/t-bastion-nsg"
    }
  }
}

mock_provider "tls" {
  mock_resource "tls_private_key" {
    defaults = {
      public_key_openssh = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCzne131kwZhZ366JqdXcJimhvLLv4+SJjB8D+4eFHpg88hdKIzf9TVkl7SVbp9VOaqLj9qn8XIOKCT1UaiTjFS0QHVWQPH8BHRn3nlz9KBU+B18/Nsunp12i8x+Oic8ahwRZrK2nLS6/OTdcUedC+rDhqHgz5cGt6bk5XiwLjmY2ColV8zUV11b9TZHsFwLnjS7spxMh/QRLwzUPIZtI2HtfBcLc4GJNKnYzv4lFOpRLzzskI9z2i8suou1CsCgXp4bBrPHJydxXHoK4TgLM+VLP9DdtExd7XWiJRjHfqGZOyG61ZKNW1Mh9Oj/m26qkbBdmYRpxwOz5Bu0fo0OhdD test"
    }
  }
}

variables {
  resource_group_name = "rg"
  location            = "centralindia"
  name_prefix         = "t"
  subnet_id           = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/v/subnets/s"
}

run "private_defaults" {
  command = plan

  assert {
    condition     = length(azurerm_public_ip.bastion) == 0 && length(azurerm_network_security_group.bastion) == 0
    error_message = "private access must not create a public IP or NSG"
  }
  assert {
    condition     = azurerm_linux_virtual_machine.bastion.admin_username == "azureuser" && azurerm_linux_virtual_machine.bastion.disable_password_authentication
    error_message = "admin defaults changed"
  }
  assert {
    condition     = tls_private_key.bastion.algorithm == "RSA" && tls_private_key.bastion.rsa_bits == 4096
    error_message = "default key must be RSA 4096"
  }
}

run "public_ip_with_allow_list" {
  variables {
    access            = "public_ip"
    allowed_ssh_cidrs = ["203.0.113.0/24"]
  }

  assert {
    condition     = length(azurerm_public_ip.bastion) == 1 && output.bastion_public_ip == azurerm_public_ip.bastion[0].ip_address
    error_message = "public_ip access must create and output a public IP"
  }
  assert {
    condition     = alltrue([for r in azurerm_network_security_group.bastion[0].security_rule : r.destination_port_range == "22" && r.source_address_prefixes == toset(["203.0.113.0/24"])])
    error_message = "NSG must allow SSH only from allowed_ssh_cidrs"
  }
  assert {
    condition     = output.bastion_ssh_private_key_pem == tls_private_key.bastion.private_key_pem
    error_message = "RSA key must be output as PEM"
  }
}

run "rejects_public_ip_without_cidrs" {
  command = plan

  variables {
    access = "public_ip"
  }

  expect_failures = [var.allowed_ssh_cidrs]
}

run "rejects_ssh_open_to_internet" {
  command = plan

  variables {
    access            = "public_ip"
    allowed_ssh_cidrs = ["0.0.0.0/0"]
  }

  expect_failures = [var.allowed_ssh_cidrs]
}
