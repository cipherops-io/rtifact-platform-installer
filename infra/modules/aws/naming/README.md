
<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
|------|---------|
| terraform | >= 1.6.0 |

## Providers

| Name | Version |
|------|---------|
| aws | n/a |

## Resources

| Name | Type |
|------|------|
| [aws_resourcegroups_group.main](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/resourcegroups_group) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| location | AWS region. | `string` | n/a | yes |
| prefix | Short name prefix for resources. | `string` | n/a | yes |
| tags | Additional resource tags. | `map(string)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| resource_group_id | n/a |
| resource_group_name | n/a |
| scope_name | n/a |
| tags | n/a |
<!-- END_TF_DOCS -->
