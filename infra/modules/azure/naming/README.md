
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
| [azurerm_resource_group.main](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/resource_group) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| create_resource_group | When false, adopt existing {prefix}-rg (e.g. after bootstrap). | `bool` | `true` | no |
| location | Azure region. | `string` | n/a | yes |
| prefix | Short name prefix for resources. | `string` | n/a | yes |
| tags | Additional resource tags. | `map(string)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| resource_group_id | n/a |
| resource_group_name | n/a |
| tags | n/a |
<!-- END_TF_DOCS -->
