
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
| enable_key_rotation | Rotate the data services KMS key automatically (kms.enable_key_rotation). | `bool` | `true` | no |
| name_prefix | n/a | `string` | n/a | yes |
| rotation_period_in_days | Days between automatic rotations (kms.rotation_period_in_days); AWS default is 365. | `number` | `null` | no |
| tags | n/a | `map(string)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| kms_key_arn | n/a |
| kms_key_id | n/a |
<!-- END_TF_DOCS -->
