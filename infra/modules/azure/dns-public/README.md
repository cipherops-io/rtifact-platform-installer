
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
| [azurerm_dns_a_record.record](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/dns_a_record) | resource |
| [azurerm_dns_zone.zone](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/dns_zone) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| resource_group_name | n/a | `string` | n/a | yes |
| tags | n/a | `map(string)` | `{}` | no |
| zones | n/a | <pre>list(object({<br/>    name = string<br/>    records = optional(list(object({<br/>      name   = string<br/>      type   = string<br/>      ttl    = optional(number)<br/>      record = string<br/>    })), [])<br/>  }))</pre> | `[]` | no |

## Outputs

| Name | Description |
|------|-------------|
| zone_ids | n/a |
<!-- END_TF_DOCS -->
