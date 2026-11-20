
<!-- BEGIN_TF_DOCS -->


## Providers

| Name | Version |
|------|---------|
| aws | n/a |
| random | n/a |

## Resources

| Name | Type |
|------|------|
| [aws_db_instance.postgresql](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/db_instance) | resource |
| [aws_db_subnet_group.postgresql](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/db_subnet_group) | resource |
| [aws_ecr_repository.main](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/ecr_repository) | resource |
| [aws_elasticache_cluster.main](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/elasticache_cluster) | resource |
| [aws_elasticache_subnet_group.redis](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/elasticache_subnet_group) | resource |
| [aws_s3_bucket.main](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/s3_bucket) | resource |
| [aws_s3_bucket_public_access_block.main](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/s3_bucket_public_access_block) | resource |
| [aws_secretsmanager_secret.workload](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/secretsmanager_secret) | resource |
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
| postgresql_sku_name | n/a | `string` | `"db.t3.medium"` | no |
| postgresql_storage_mb | n/a | `number` | `1048576` | no |
| postgresql_version | n/a | `string` | `"16.4"` | no |
| private_endpoint_subnet_id | n/a | `string` | n/a | yes |
| redis_sku_name | n/a | `string` | `"cache.t3.small"` | no |
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
