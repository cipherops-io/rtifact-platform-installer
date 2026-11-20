
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
| tags | n/a | `map(string)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| bastion_host_id | n/a |
<!-- END_TF_DOCS -->
