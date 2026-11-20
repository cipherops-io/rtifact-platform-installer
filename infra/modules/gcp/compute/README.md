
<!-- BEGIN_TF_DOCS -->


## Providers

| Name | Version |
|------|---------|
| google | n/a |
| tls | n/a |

## Resources

| Name | Type |
|------|------|
| [google_compute_instance.bastion](https://registry.terraform.io/providers/opentofu/google/latest/docs/resources/compute_instance) | resource |
| [tls_private_key.bastion](https://registry.terraform.io/providers/opentofu/tls/latest/docs/resources/private_key) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| location | n/a | `string` | n/a | yes |
| name_prefix | n/a | `string` | n/a | yes |
| os_disk_size_gb | n/a | `number` | n/a | yes |
| project_id | n/a | `string` | n/a | yes |
| scope_name | n/a | `string` | n/a | yes |
| subnet_id | n/a | `string` | n/a | yes |
| tags | n/a | `map(string)` | `{}` | no |
| vm_size | n/a | `string` | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| bastion_private_ip | n/a |
| bastion_ssh_private_key_pem | n/a |
| bastion_vm_id | n/a |
<!-- END_TF_DOCS -->
