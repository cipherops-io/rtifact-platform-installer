
<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
|------|---------|
| terraform | >= 1.6.0 |

## Providers

| Name | Version |
|------|---------|
| azurerm | n/a |

## Resources

| Name | Type |
|------|------|
| [azurerm_bastion_host.main](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/bastion_host) | resource |
| [azurerm_public_ip.bastion](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/public_ip) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| bastion_subnet_id | n/a | `string` | n/a | yes |
| location | n/a | `string` | n/a | yes |
| name_prefix | n/a | `string` | n/a | yes |
| resource_group_name | n/a | `string` | n/a | yes |
| settings | Azure Bastion host settings (bastion.azure_bastion in customer YAML). tunneling_enabled is required for `az network bastion ssh`. | <pre>object({<br/>    sku                = optional(string, "Standard")<br/>    scale_units        = optional(number, 2)<br/>    tunneling_enabled  = optional(bool, true)<br/>    copy_paste_enabled = optional(bool, true)<br/>    file_copy_enabled  = optional(bool, false)<br/>    ip_connect_enabled = optional(bool, false)<br/>  })</pre> | `{}` | no |
| tags | n/a | `map(string)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| bastion_host_id | n/a |
| bastion_host_name | n/a |
<!-- END_TF_DOCS -->
