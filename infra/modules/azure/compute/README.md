
<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
|------|---------|
| terraform | >= 1.6.0 |
| tls | >= 4.0 |

## Providers

| Name | Version |
|------|---------|
| azurerm | n/a |
| tls | >= 4.0 |

## Resources

| Name | Type |
|------|------|
| [azurerm_linux_virtual_machine.bastion](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/linux_virtual_machine) | resource |
| [azurerm_network_interface.bastion](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/network_interface) | resource |
| [azurerm_network_interface_security_group_association.bastion](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/network_interface_security_group_association) | resource |
| [azurerm_network_security_group.bastion](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/network_security_group) | resource |
| [azurerm_public_ip.bastion](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/public_ip) | resource |
| [tls_private_key.bastion](https://registry.terraform.io/providers/opentofu/tls/latest/docs/resources/private_key) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| access | How operators reach the bastion VM: private (VNet only), public_ip (public IP plus NSG allow-list) or azure_bastion (Azure Bastion tunnel). | `string` | `"private"` | no |
| admin_username | n/a | `string` | `"azureuser"` | no |
| allowed_ssh_cidrs | Source CIDRs allowed to SSH to the bastion public IP (bastion.allowed_ssh_cidrs, public_ip access only). | `list(string)` | `[]` | no |
| image | Marketplace image for the bastion VM (bastion.image). | <pre>object({<br/>    publisher = optional(string, "Canonical")<br/>    offer     = optional(string, "0001-com-ubuntu-server-jammy")<br/>    sku       = optional(string, "22_04-lts-gen2")<br/>    version   = optional(string, "latest")<br/>  })</pre> | `{}` | no |
| location | n/a | `string` | n/a | yes |
| name_prefix | n/a | `string` | n/a | yes |
| os_disk_caching | n/a | `string` | `"ReadWrite"` | no |
| os_disk_size_gb | n/a | `number` | `100` | no |
| os_disk_type | n/a | `string` | `"Premium_LRS"` | no |
| private_ip | Private IP of the bastion NIC in the management subnet (bastion.private_ip). | <pre>object({<br/>    allocation_method = optional(string, "Dynamic")<br/>    address           = optional(string)<br/>  })</pre> | `{}` | no |
| resource_group_name | n/a | `string` | n/a | yes |
| ssh_key | Generated SSH key for the bastion admin user (bastion.ssh_key). Azure Linux VMs accept RSA (2048+) and ED25519. | <pre>object({<br/>    algorithm = optional(string, "RSA")<br/>    rsa_bits  = optional(number, 4096)<br/>  })</pre> | `{}` | no |
| subnet_id | n/a | `string` | n/a | yes |
| tags | n/a | `map(string)` | `{}` | no |
| vm_size | n/a | `string` | `"Standard_D4s_v5"` | no |

## Outputs

| Name | Description |
|------|-------------|
| bastion_admin_username | n/a |
| bastion_private_ip | n/a |
| bastion_public_ip | n/a |
| bastion_ssh_private_key_pem | n/a |
| bastion_vm_id | n/a |
<!-- END_TF_DOCS -->
