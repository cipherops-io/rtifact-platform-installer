resource "aws_secretsmanager_secret" "workload" {
  count = var.features.postgresql_flexible_server ? 1 : 0

  name = "${var.name_prefix}-workload-secrets"
  tags = var.tags
}