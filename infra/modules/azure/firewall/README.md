
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
| [azurerm_firewall.main](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/firewall) | resource |
| [azurerm_firewall_policy.main](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/firewall_policy) | resource |
| [azurerm_public_ip.firewall](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/public_ip) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| firewall_subnet_id | n/a | `string` | n/a | yes |
| location | n/a | `string` | n/a | yes |
| name_prefix | n/a | `string` | n/a | yes |
| resource_group_name | n/a | `string` | n/a | yes |
| sku_tier | n/a | `string` | `"Standard"` | no |
| tags | n/a | `map(string)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| firewall_id | n/a |
| firewall_policy_id | n/a |
<!-- END_TF_DOCS -->
