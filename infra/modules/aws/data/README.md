
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
| [aws_elasticache_replication_group.redis](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/elasticache_replication_group) | resource |
| [aws_elasticache_subnet_group.redis](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/elasticache_subnet_group) | resource |
| [aws_s3_bucket.main](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/s3_bucket) | resource |
| [aws_s3_bucket_ownership_controls.main](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/s3_bucket_ownership_controls) | resource |
| [aws_s3_bucket_public_access_block.main](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/s3_bucket_public_access_block) | resource |
| [aws_s3_bucket_server_side_encryption_configuration.main](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/s3_bucket_server_side_encryption_configuration) | resource |
| [aws_s3_bucket_versioning.main](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/s3_bucket_versioning) | resource |
| [aws_security_group.postgresql](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/security_group) | resource |
| [aws_security_group.redis](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/security_group) | resource |
| [aws_vpc_security_group_ingress_rule.postgresql](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/vpc_security_group_ingress_rule) | resource |
| [aws_vpc_security_group_ingress_rule.redis](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/vpc_security_group_ingress_rule) | resource |
| [random_string.suffix](https://registry.terraform.io/providers/opentofu/random/latest/docs/resources/string) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| container_registry | n/a | <pre>object({<br/>    image_tag_mutability = optional(string, "IMMUTABLE")<br/>    scan_on_push         = optional(bool, true)<br/>  })</pre> | `{}` | no |
| features | n/a | <pre>object({<br/>    postgresql         = bool<br/>    redis              = bool<br/>    object_storage     = bool<br/>    container_registry = bool<br/>  })</pre> | n/a | yes |
| kms_key_arn | Customer-managed key for encryption; AWS-managed keys are used when null. | `string` | `null` | no |
| name_prefix | n/a | `string` | n/a | yes |
| object_storage | n/a | <pre>object({<br/>    versioning          = optional(bool, true)<br/>    object_ownership    = optional(string, "BucketOwnerEnforced")<br/>    block_public_access = optional(bool, true)<br/>  })</pre> | `{}` | no |
| postgresql | n/a | <pre>object({<br/>    engine_version           = optional(string, "16")<br/>    instance_class           = optional(string, "db.t4g.medium")<br/>    allocated_storage_gb     = optional(number, 100)<br/>    max_allocated_storage_gb = optional(number, 500)<br/>    storage_type             = optional(string, "gp3")<br/>    multi_az                 = optional(bool, false)<br/>    backup_retention_days    = optional(number, 7)<br/>    deletion_protection      = optional(bool, true)<br/>    subnet_tier              = optional(string, "datastore")<br/>    port                     = optional(number, 5432)<br/>    db_name                  = optional(string, "aios")<br/>    username                 = optional(string, "aiosadmin")<br/>    publicly_accessible      = optional(bool, false)<br/>    copy_tags_to_snapshot    = optional(bool, true)<br/>    skip_final_snapshot      = optional(bool, false)<br/>  })</pre> | `{}` | no |
| redis | n/a | <pre>object({<br/>    engine                     = optional(string, "redis")<br/>    engine_version             = optional(string, "7.1")<br/>    node_type                  = optional(string, "cache.t4g.small")<br/>    num_cache_clusters         = optional(number, 1)<br/>    subnet_tier                = optional(string, "datastore")<br/>    port                       = optional(number, 6379)<br/>    at_rest_encryption_enabled = optional(bool, true)<br/>    transit_encryption_enabled = optional(bool, true)<br/>  })</pre> | `{}` | no |
| subnet_ids_by_tier | n/a | `map(list(string))` | n/a | yes |
| tags | n/a | `map(string)` | `{}` | no |
| vpc_cidrs | CIDRs allowed to reach PostgreSQL and Redis. | `list(string)` | n/a | yes |
| vpc_id | n/a | `string` | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| container_registry_arn | n/a |
| container_registry_url | n/a |
| object_storage_bucket_arn | n/a |
| object_storage_bucket_name | n/a |
| postgresql_endpoint | n/a |
| postgresql_master_user_secret_arn | n/a |
| postgresql_port | n/a |
| redis_primary_endpoint | n/a |
<!-- END_TF_DOCS -->
