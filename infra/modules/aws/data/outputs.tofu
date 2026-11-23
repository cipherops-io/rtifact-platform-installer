output "postgresql_endpoint" {
  value = try(aws_db_instance.postgresql[0].address, null)
}

output "postgresql_port" {
  value = try(aws_db_instance.postgresql[0].port, null)
}

output "postgresql_master_user_secret_arn" {
  value = try(aws_db_instance.postgresql[0].master_user_secret[0].secret_arn, null)
}

output "redis_primary_endpoint" {
  value = try(aws_elasticache_replication_group.redis[0].primary_endpoint_address, null)
}

output "object_storage_bucket_name" {
  value = try(aws_s3_bucket.main[0].id, null)
}

output "object_storage_bucket_arn" {
  value = try(aws_s3_bucket.main[0].arn, null)
}

output "container_registry_url" {
  value = try(aws_ecr_repository.main[0].repository_url, null)
}

output "container_registry_arn" {
  value = try(aws_ecr_repository.main[0].arn, null)
}
