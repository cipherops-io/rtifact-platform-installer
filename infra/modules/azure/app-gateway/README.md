
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
| [azurerm_application_gateway.main](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/application_gateway) | resource |
| [azurerm_public_ip.appgw](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/public_ip) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| availability_zones | n/a | `list(string)` | `[]` | no |
| capacity | n/a | `number` | `2` | no |
| location | n/a | `string` | n/a | yes |
| name_prefix | n/a | `string` | n/a | yes |
| resource_group_name | n/a | `string` | n/a | yes |
| sku_name | n/a | `string` | `"WAF_v2"` | no |
| sku_tier | n/a | `string` | `"WAF_v2"` | no |
| subnet_id | n/a | `string` | n/a | yes |
| tags | n/a | `map(string)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| application_gateway_id | n/a |
| public_ip | n/a |
<!-- END_TF_DOCS -->
