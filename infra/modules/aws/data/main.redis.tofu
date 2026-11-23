resource "aws_elasticache_subnet_group" "redis" {
  count = var.features.redis ? 1 : 0

  name       = "${local.base_name}-redis"
  subnet_ids = var.subnet_ids_by_tier[var.redis.subnet_tier]
  tags       = var.tags
}

resource "aws_security_group" "redis" {
  count = var.features.redis ? 1 : 0

  name        = "${local.base_name}-redis"
  description = "Redis from inside the VPC"
  vpc_id      = var.vpc_id
  tags        = merge(var.tags, { Name = "${local.base_name}-redis" })
}

resource "aws_vpc_security_group_ingress_rule" "redis" {
  for_each = var.features.redis ? toset(var.vpc_cidrs) : toset([])

  security_group_id = aws_security_group.redis[0].id
  cidr_ipv4         = each.value
  ip_protocol       = "tcp"
  from_port         = 6379
  to_port           = 6379
}

resource "aws_elasticache_replication_group" "redis" {
  count = var.features.redis ? 1 : 0

  replication_group_id       = "${local.base_name}-redis"
  description                = "${var.name_prefix} Redis"
  engine                     = "redis"
  engine_version             = var.redis.engine_version
  node_type                  = var.redis.node_type
  num_cache_clusters         = var.redis.num_cache_clusters
  automatic_failover_enabled = var.redis.num_cache_clusters > 1
  multi_az_enabled           = var.redis.num_cache_clusters > 1
  port                       = 6379
  subnet_group_name          = aws_elasticache_subnet_group.redis[0].name
  security_group_ids         = [aws_security_group.redis[0].id]
  at_rest_encryption_enabled = true
  transit_encryption_enabled = true
  kms_key_id                 = var.kms_key_arn
  tags                       = var.tags
}
