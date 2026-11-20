
<!-- BEGIN_TF_DOCS -->


## Providers

| Name | Version |
|------|---------|
| google | n/a |
| random | n/a |

## Resources

| Name | Type |
|------|------|
| [google_artifact_registry_repository.main](https://registry.terraform.io/providers/opentofu/google/latest/docs/resources/artifact_registry_repository) | resource |
| [google_redis_instance.main](https://registry.terraform.io/providers/opentofu/google/latest/docs/resources/redis_instance) | resource |
| [google_secret_manager_secret.workload](https://registry.terraform.io/providers/opentofu/google/latest/docs/resources/secret_manager_secret) | resource |
| [google_sql_database_instance.postgresql](https://registry.terraform.io/providers/opentofu/google/latest/docs/resources/sql_database_instance) | resource |
| [google_storage_bucket.main](https://registry.terraform.io/providers/opentofu/google/latest/docs/resources/storage_bucket) | resource |
| [random_password.pg](https://registry.terraform.io/providers/opentofu/random/latest/docs/resources/password) | resource |
| [random_string.suffix](https://registry.terraform.io/providers/opentofu/random/latest/docs/resources/string) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| acr_sku | n/a | `string` | `"private"` | no |
| dns_scope_name | n/a | `string` | n/a | yes |
| features | n/a | <pre>object({<br/>    postgresql_flexible_server = bool<br/>    container_registry         = bool<br/>    storage_account            = bool<br/>    redis_cache                = bool<br/>    private_dns_zones          = bool<br/>  })</pre> | n/a | yes |
| location | n/a | `string` | n/a | yes |
| name_prefix | n/a | `string` | n/a | yes |
| node_subnet_id | n/a | `string` | n/a | yes |
| postgresql_sku_name | n/a | `string` | `"db-custom-2-8192"` | no |
| postgresql_storage_mb | n/a | `number` | `1048576` | no |
| postgresql_version | n/a | `string` | `"POSTGRES_16"` | no |
| private_endpoint_subnet_id | n/a | `string` | n/a | yes |
| project_id | n/a | `string` | n/a | yes |
| redis_sku_name | n/a | `string` | `"STANDARD_HA"` | no |
| scope_name | n/a | `string` | n/a | yes |
| storage_backup_gb | n/a | `number` | `250` | no |
| storage_corpus_gb | n/a | `number` | `250` | no |
| tags | n/a | `map(string)` | `{}` | no |
| vnet_id | n/a | `string` | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| acr_id | n/a |
| postgresql_id | n/a |
| private_dns_zone_ids | n/a |
| redis_id | n/a |
| storage_account_id | n/a |
<!-- END_TF_DOCS -->
