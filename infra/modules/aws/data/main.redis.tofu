resource "aws_elasticache_subnet_group" "redis" {
  count = var.features.redis_cache ? 1 : 0

  name       = "${var.name_prefix}-redis"
  subnet_ids = distinct([var.node_subnet_id, var.private_endpoint_subnet_id])
}

resource "aws_elasticache_cluster" "main" {
  count = var.features.redis_cache ? 1 : 0

  cluster_id           = "${var.name_prefix}-redis"
  engine               = "redis"
  node_type            = var.redis_sku_name
  num_cache_nodes      = 1
  parameter_group_name = "default.redis7"
  subnet_group_name    = aws_elasticache_subnet_group.redis[0].name
  tags                 = var.tags
}