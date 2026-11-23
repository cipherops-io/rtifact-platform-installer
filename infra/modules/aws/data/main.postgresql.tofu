resource "aws_db_subnet_group" "postgresql" {
  count = var.features.postgresql ? 1 : 0

  name       = "${local.base_name}-pg"
  subnet_ids = var.subnet_ids_by_tier[var.postgresql.subnet_tier]
  tags       = var.tags
}

resource "aws_security_group" "postgresql" {
  count = var.features.postgresql ? 1 : 0

  name        = "${local.base_name}-pg"
  description = "PostgreSQL from inside the VPC"
  vpc_id      = var.vpc_id
  tags        = merge(var.tags, { Name = "${local.base_name}-pg" })
}

resource "aws_vpc_security_group_ingress_rule" "postgresql" {
  for_each = var.features.postgresql ? toset(var.vpc_cidrs) : toset([])

  security_group_id = aws_security_group.postgresql[0].id
  cidr_ipv4         = each.value
  ip_protocol       = "tcp"
  from_port         = var.postgresql.port
  to_port           = var.postgresql.port
}

resource "aws_db_instance" "postgresql" {
  count = var.features.postgresql ? 1 : 0

  identifier                    = "${local.base_name}-pg"
  engine                        = "postgres"
  engine_version                = var.postgresql.engine_version
  instance_class                = var.postgresql.instance_class
  allocated_storage             = var.postgresql.allocated_storage_gb
  max_allocated_storage         = var.postgresql.max_allocated_storage_gb
  storage_type                  = var.postgresql.storage_type
  storage_encrypted             = true
  kms_key_id                    = var.kms_key_arn
  db_name                       = var.postgresql.db_name
  username                      = var.postgresql.username
  port                          = var.postgresql.port
  manage_master_user_password   = true
  master_user_secret_kms_key_id = var.kms_key_arn
  db_subnet_group_name          = aws_db_subnet_group.postgresql[0].name
  vpc_security_group_ids        = [aws_security_group.postgresql[0].id]
  publicly_accessible           = var.postgresql.publicly_accessible
  multi_az                      = var.postgresql.multi_az
  backup_retention_period       = var.postgresql.backup_retention_days
  deletion_protection           = var.postgresql.deletion_protection
  copy_tags_to_snapshot         = var.postgresql.copy_tags_to_snapshot
  skip_final_snapshot           = var.postgresql.skip_final_snapshot
  final_snapshot_identifier     = var.postgresql.skip_final_snapshot ? null : "${local.base_name}-pg-final"
  tags                          = var.tags
}
