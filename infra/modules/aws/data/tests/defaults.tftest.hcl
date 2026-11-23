mock_provider "aws" {}

mock_provider "random" {}

variables {
  name_prefix        = "t"
  features           = { postgresql = true, redis = true, object_storage = true, container_registry = false }
  vpc_id             = "vpc-0123456789abcdef0"
  vpc_cidrs          = ["10.0.0.0/16"]
  subnet_ids_by_tier = { datastore = ["subnet-a", "subnet-b"] }
}

run "defaults_match_previous_literals" {
  command = plan

  assert {
    condition = (
      aws_db_instance.postgresql[0].engine == "postgres" &&
      aws_db_instance.postgresql[0].storage_type == "gp3" &&
      aws_db_instance.postgresql[0].db_name == "aios" &&
      aws_db_instance.postgresql[0].username == "aiosadmin" &&
      aws_db_instance.postgresql[0].port == 5432 &&
      !aws_db_instance.postgresql[0].publicly_accessible &&
      aws_db_instance.postgresql[0].copy_tags_to_snapshot &&
      !aws_db_instance.postgresql[0].skip_final_snapshot
    )
    error_message = "PostgreSQL defaults changed"
  }
  assert {
    condition     = alltrue([for r in aws_vpc_security_group_ingress_rule.postgresql : r.from_port == 5432 && r.to_port == 5432])
    error_message = "PostgreSQL ingress must follow postgresql.port"
  }
  assert {
    condition = (
      aws_elasticache_replication_group.redis[0].engine == "redis" &&
      aws_elasticache_replication_group.redis[0].port == 6379 &&
      aws_elasticache_replication_group.redis[0].at_rest_encryption_enabled &&
      aws_elasticache_replication_group.redis[0].transit_encryption_enabled
    )
    error_message = "Redis defaults changed"
  }
  assert {
    condition = (
      aws_s3_bucket_ownership_controls.main[0].rule[0].object_ownership == "BucketOwnerEnforced" &&
      aws_s3_bucket_public_access_block.main[0].block_public_acls &&
      aws_s3_bucket_public_access_block.main[0].restrict_public_buckets
    )
    error_message = "object storage defaults changed"
  }
}

run "rejects_invalid_storage_type" {
  command = plan

  variables {
    postgresql = { storage_type = "magnetic" }
  }

  expect_failures = [var.postgresql]
}
