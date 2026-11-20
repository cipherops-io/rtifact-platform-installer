
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
| [tls_private_key.bastion](https://registry.terraform.io/providers/opentofu/tls/latest/docs/resources/private_key) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| location | n/a | `string` | n/a | yes |
| name_prefix | n/a | `string` | n/a | yes |
| os_disk_size_gb | n/a | `number` | `100` | no |
| os_disk_type | n/a | `string` | `"Premium_LRS"` | no |
| resource_group_name | n/a | `string` | n/a | yes |
| subnet_id | n/a | `string` | n/a | yes |
| tags | n/a | `map(string)` | `{}` | no |
| vm_size | n/a | `string` | `"Standard_D4s_v5"` | no |

## Outputs

| Name | Description |
|------|-------------|
| bastion_private_ip | n/a |
| bastion_ssh_private_key_pem | n/a |
| bastion_vm_id | n/a |
<!-- END_TF_DOCS -->
