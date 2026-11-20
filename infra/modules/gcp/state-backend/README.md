
<!-- BEGIN_TF_DOCS -->


## Providers

| Name | Version |
|------|---------|
| google | n/a |

## Resources

| Name | Type |
|------|------|
| [google_storage_bucket.tfstate](https://registry.terraform.io/providers/opentofu/google/latest/docs/resources/storage_bucket) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| location | n/a | `string` | n/a | yes |
| name_prefix | n/a | `string` | n/a | yes |
| project_id | n/a | `string` | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| gcs_bucket_name | n/a |
| resource_group_name | Compatibility alias for installer discovery. |
<!-- END_TF_DOCS -->
