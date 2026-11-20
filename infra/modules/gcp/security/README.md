
<!-- BEGIN_TF_DOCS -->


## Providers

| Name | Version |
|------|---------|
| google | n/a |

## Resources

| Name | Type |
|------|------|
| [google_kms_crypto_key.main](https://registry.terraform.io/providers/opentofu/google/latest/docs/resources/kms_crypto_key) | resource |
| [google_kms_key_ring.main](https://registry.terraform.io/providers/opentofu/google/latest/docs/resources/kms_key_ring) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| location | n/a | `string` | n/a | yes |
| name_prefix | n/a | `string` | n/a | yes |
| scope_name | n/a | `string` | n/a | yes |
| sku | n/a | `string` | `"standard"` | no |
| tags | n/a | `map(string)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| key_vault_id | n/a |
| key_vault_uri | n/a |
<!-- END_TF_DOCS -->
