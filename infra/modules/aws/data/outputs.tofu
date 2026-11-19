output "postgresql_id" {
  value = try(aws_db_instance.postgresql[0].id, null)
}

output "storage_account_id" {
  value = try(aws_s3_bucket.main[0].id, null)
}

output "acr_id" {
  value = try(aws_ecr_repository.main[0].arn, null)
}

output "redis_id" {
  value = try(aws_elasticache_cluster.main[0].id, null)
}

output "private_dns_zone_ids" {
  value = {}
}