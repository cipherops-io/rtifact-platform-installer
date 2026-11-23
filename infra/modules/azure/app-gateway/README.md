
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
| backend_http_settings | Placeholder backend HTTP settings (application_gateway.backend_http_settings); AGIC replaces them for real routes. | <pre>object({<br/>    port                  = optional(number, 80)<br/>    protocol              = optional(string, "Http")<br/>    path                  = optional(string, "/")<br/>    cookie_based_affinity = optional(string, "Disabled")<br/>    request_timeout       = optional(number, 30)<br/>  })</pre> | `{}` | no |
| capacity | n/a | `number` | `2` | no |
| frontend_port | Listener port of the placeholder HTTP listener (application_gateway.frontend_port). | `number` | `80` | no |
| location | n/a | `string` | n/a | yes |
| name_prefix | n/a | `string` | n/a | yes |
| resource_group_name | n/a | `string` | n/a | yes |
| routing_rule_priority | Priority of the default request routing rule (application_gateway.routing_rule_priority). | `number` | `1` | no |
| sku_name | n/a | `string` | `"WAF_v2"` | no |
| sku_tier | n/a | `string` | `"WAF_v2"` | no |
| ssl_policy | Predefined TLS policy (application_gateway.ssl_policy). See https://learn.microsoft.com/azure/application-gateway/application-gateway-ssl-policy-overview. | <pre>object({<br/>    policy_type = optional(string, "Predefined")<br/>    policy_name = optional(string, "AppGwSslPolicy20170401S")<br/>  })</pre> | `{}` | no |
| subnet_id | n/a | `string` | n/a | yes |
| tags | n/a | `map(string)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| application_gateway_id | n/a |
| public_ip | n/a |
<!-- END_TF_DOCS -->
