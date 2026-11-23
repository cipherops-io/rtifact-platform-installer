
<!-- BEGIN_TF_DOCS -->


## Providers

| Name | Version |
|------|---------|
| aws | n/a |

## Resources

| Name | Type |
|------|------|
| [aws_kms_alias.main](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/kms_alias) | resource |
| [aws_kms_key.main](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/kms_key) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| deletion_window_days | n/a | `number` | `30` | no |
| name_prefix | n/a | `string` | n/a | yes |
| tags | n/a | `map(string)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| kms_key_arn | n/a |
| kms_key_id | n/a |
<!-- END_TF_DOCS -->
