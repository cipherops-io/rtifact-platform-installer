
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
| [azurerm_private_endpoint.main](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/private_endpoint) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| connection | Private service connection approval settings (network.private_endpoint in customer YAML). | <pre>object({<br/>    is_manual_connection = optional(bool, false)<br/>    request_message      = optional(string)<br/>  })</pre> | `{}` | no |
| location | n/a | `string` | n/a | yes |
| name | n/a | `string` | n/a | yes |
| private_connection_resource_id | n/a | `string` | n/a | yes |
| private_dns_zone_ids | n/a | `list(string)` | n/a | yes |
| resource_group_name | n/a | `string` | n/a | yes |
| subnet_id | n/a | `string` | n/a | yes |
| subresource_names | n/a | `list(string)` | n/a | yes |
| tags | n/a | `map(string)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| id | n/a |
<!-- END_TF_DOCS -->
