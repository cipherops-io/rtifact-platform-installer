
<!-- BEGIN_TF_DOCS -->


## Providers

| Name | Version |
|------|---------|
| google | n/a |

## Resources

| Name | Type |
|------|------|
| [google_compute_forwarding_rule.psc](https://registry.terraform.io/providers/opentofu/google/latest/docs/resources/compute_forwarding_rule) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| location | n/a | `string` | n/a | yes |
| name | n/a | `string` | n/a | yes |
| network_id | n/a | `string` | n/a | yes |
| scope_name | n/a | `string` | n/a | yes |
| subnet_id | n/a | `string` | n/a | yes |
| tags | n/a | `map(string)` | `{}` | no |
| target_service | n/a | `string` | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| endpoint_id | n/a |
<!-- END_TF_DOCS -->
