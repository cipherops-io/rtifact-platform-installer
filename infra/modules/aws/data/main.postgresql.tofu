resource "random_password" "pg" {
  count = var.features.postgresql_flexible_server ? 1 : 0

  length  = 24
  special = true
}

resource "aws_db_subnet_group" "postgresql" {
  count = var.features.postgresql_flexible_server ? 1 : 0

  name       = "${var.name_prefix}-pg"
  subnet_ids = distinct([var.node_subnet_id, var.private_endpoint_subnet_id])
  tags       = var.tags
}

resource "aws_db_instance" "postgresql" {
  count = var.features.postgresql_flexible_server ? 1 : 0

  identifier           = "${var.name_prefix}-pg"
  engine               = "postgres"
  engine_version       = var.postgresql_version
  instance_class       = var.postgresql_sku_name
  allocated_storage    = max(floor(var.postgresql_storage_mb / 1024), 20)
  db_name              = "aios"
  username             = "aiosadmin"
  password             = random_password.pg[0].result
  db_subnet_group_name = aws_db_subnet_group.postgresql[0].name
  publicly_accessible  = false
  skip_final_snapshot  = true
  tags                 = var.tags
}