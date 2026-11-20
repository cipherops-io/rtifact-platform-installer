
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
| [azurerm_public_ip.er](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/public_ip) | resource |
| [azurerm_public_ip.vpn](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/public_ip) | resource |
| [azurerm_virtual_network_gateway.express_route](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/virtual_network_gateway) | resource |
| [azurerm_virtual_network_gateway.vpn](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/virtual_network_gateway) | resource |
| [azurerm_virtual_network_gateway_connection.er](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/virtual_network_gateway_connection) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| express_route | n/a | <pre>object({<br/>    sku                      = optional(string, "ErGw2AZ")<br/>    express_route_circuit_id = optional(string)<br/>  })</pre> | <pre>{<br/>  "sku": "ErGw2AZ"<br/>}</pre> | no |
| express_route_enabled | n/a | `bool` | `false` | no |
| gateway_subnet_id | n/a | `string` | n/a | yes |
| location | n/a | `string` | n/a | yes |
| name_prefix | n/a | `string` | n/a | yes |
| resource_group_name | n/a | `string` | n/a | yes |
| tags | n/a | `map(string)` | `{}` | no |
| vpn | n/a | <pre>object({<br/>    sku           = optional(string, "VpnGw2")<br/>    gateway_type  = optional(string, "Vpn")<br/>    vpn_type      = optional(string, "RouteBased")<br/>    active_active = optional(bool, false)<br/>  })</pre> | <pre>{<br/>  "active_active": false,<br/>  "gateway_type": "Vpn",<br/>  "sku": "VpnGw2",<br/>  "vpn_type": "RouteBased"<br/>}</pre> | no |
| vpn_enabled | n/a | `bool` | `false` | no |

## Outputs

| Name | Description |
|------|-------------|
| express_route_gateway_id | n/a |
| vpn_gateway_id | n/a |
<!-- END_TF_DOCS -->
