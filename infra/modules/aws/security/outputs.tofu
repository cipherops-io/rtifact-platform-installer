output "key_vault_id" {
  value = aws_kms_key.main.id
}

output "key_vault_uri" {
  value = aws_kms_key.main.arn
}
