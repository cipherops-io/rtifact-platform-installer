output "s3_bucket_name" {
  value = module.state_backend.s3_bucket_name
}

output "dynamodb_table_name" {
  value = module.state_backend.dynamodb_table_name
}
