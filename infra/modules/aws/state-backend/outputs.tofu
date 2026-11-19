output "s3_bucket_name" {
  value = aws_s3_bucket.tfstate.id
}

output "dynamodb_table_name" {
  value = aws_dynamodb_table.tfstate_lock.name
}

output "resource_group_name" {
  description = "Compatibility alias for installer discovery (scope label)."
  value       = "${var.name_prefix}-tfstate"
}
