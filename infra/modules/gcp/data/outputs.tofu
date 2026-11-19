output "postgresql_id" {
  value = try(google_sql_database_instance.postgresql[0].id, null)
}

output "storage_account_id" {
  value = try(google_storage_bucket.main[0].id, null)
}

output "acr_id" {
  value = try(google_artifact_registry_repository.main[0].id, null)
}

output "redis_id" {
  value = try(google_redis_instance.main[0].id, null)
}

output "private_dns_zone_ids" {
  value = {}
}