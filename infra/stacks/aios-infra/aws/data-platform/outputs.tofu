output "kms_key_arn" {
  value = one(module.security[*].kms_key_arn)
}

output "postgresql_endpoint" {
  value = one(module.data[*].postgresql_endpoint)
}

output "postgresql_port" {
  value = one(module.data[*].postgresql_port)
}

output "postgresql_master_user_secret_arn" {
  value = one(module.data[*].postgresql_master_user_secret_arn)
}

output "redis_primary_endpoint" {
  value = one(module.data[*].redis_primary_endpoint)
}

output "object_storage_bucket_name" {
  value = one(module.data[*].object_storage_bucket_name)
}

output "object_storage_bucket_arn" {
  value = one(module.data[*].object_storage_bucket_arn)
}

output "container_registry_url" {
  value = one(module.data[*].container_registry_url)
}

output "container_registry_arn" {
  value = one(module.data[*].container_registry_arn)
}
