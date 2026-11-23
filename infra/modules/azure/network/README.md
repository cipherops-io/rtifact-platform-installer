
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
| [azurerm_nat_gateway.main](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/nat_gateway) | resource |
| [azurerm_nat_gateway_public_ip_association.main](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/nat_gateway_public_ip_association) | resource |
| [azurerm_private_dns_resolver.main](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/private_dns_resolver) | resource |
| [azurerm_private_dns_resolver_inbound_endpoint.main](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/private_dns_resolver_inbound_endpoint) | resource |
| [azurerm_private_dns_resolver_outbound_endpoint.main](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/private_dns_resolver_outbound_endpoint) | resource |
| [azurerm_public_ip.nat](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/public_ip) | resource |
| [azurerm_subnet.bastion](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/subnet) | resource |
| [azurerm_subnet.dns_resolver_inbound](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/subnet) | resource |
| [azurerm_subnet.dns_resolver_outbound](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/subnet) | resource |
| [azurerm_subnet.firewall](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/subnet) | resource |
| [azurerm_subnet.gateway](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/subnet) | resource |
| [azurerm_subnet.load_balancer](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/subnet) | resource |
| [azurerm_subnet.management](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/subnet) | resource |
| [azurerm_subnet.node](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/subnet) | resource |
| [azurerm_subnet.private_endpoint](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/subnet) | resource |
| [azurerm_subnet_nat_gateway_association.node](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/subnet_nat_gateway_association) | resource |
| [azurerm_virtual_network.main](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/virtual_network) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| azure_bastion_subnet_cidr | n/a | `string` | `null` | no |
| azure_firewall_subnet_cidr | n/a | `string` | `null` | no |
| bastion_managed | n/a | `bool` | `false` | no |
| dns_resolver_inbound | Inbound endpoint IP configuration for the private DNS resolver (private_dns_resolver in customer YAML). | <pre>object({<br/>    enabled                      = optional(bool, true)<br/>    private_ip_allocation_method = optional(string, "Dynamic")<br/>    private_ip_address           = optional(string)<br/>  })</pre> | `{}` | no |
| dns_resolver_inbound_subnet_cidr | n/a | `string` | `null` | no |
| dns_resolver_outbound_enabled | Create the private DNS resolver outbound endpoint (private_dns_resolver.outbound_endpoint_enabled). | `bool` | `true` | no |
| dns_resolver_outbound_subnet_cidr | n/a | `string` | `null` | no |
| features | n/a | <pre>object({<br/>    azure_firewall              = optional(bool, false)<br/>    azure_vpn_gateway           = optional(bool, false)<br/>    azure_express_route_gateway = optional(bool, false)<br/>    private_dns_resolver        = optional(bool, false)<br/>  })</pre> | <pre>{<br/>  "azure_express_route_gateway": false,<br/>  "azure_firewall": false,<br/>  "azure_vpn_gateway": false,<br/>  "private_dns_resolver": false<br/>}</pre> | no |
| gateway_subnet_cidr | n/a | `string` | `null` | no |
| load_balancer_subnet_cidr | n/a | `string` | n/a | yes |
| location | n/a | `string` | n/a | yes |
| management_subnet_cidr | n/a | `string` | n/a | yes |
| name_prefix | n/a | `string` | n/a | yes |
| nat | NAT gateway for the node subnet and its public IP (network.nat in customer YAML). | <pre>object({<br/>    sku_name = optional(string, "Standard")<br/>    public_ip = optional(object({<br/>      sku               = optional(string, "Standard")<br/>      allocation_method = optional(string, "Static")<br/>    }), {})<br/>  })</pre> | `{}` | no |
| node_subnet_cidr | n/a | `string` | n/a | yes |
| private_endpoint_subnet_cidr | n/a | `string` | n/a | yes |
| resource_group_name | n/a | `string` | n/a | yes |
| tags | n/a | `map(string)` | `{}` | no |
| vnet_cidr | n/a | `string` | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| bastion_subnet_id | n/a |
| firewall_subnet_id | n/a |
| gateway_subnet_id | n/a |
| load_balancer_subnet_id | n/a |
| management_subnet_id | n/a |
| nat_gateway_id | n/a |
| node_subnet_id | n/a |
| private_dns_resolver_id | n/a |
| private_endpoint_subnet_id | n/a |
| vnet_id | n/a |
| vnet_name | n/a |
<!-- END_TF_DOCS -->
