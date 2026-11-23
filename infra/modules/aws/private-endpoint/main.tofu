resource "aws_vpc_endpoint" "main" {
  vpc_id              = var.vpc_id
  service_name        = var.service_name
  vpc_endpoint_type   = "Interface"
  subnet_ids          = [var.subnet_id]
  private_dns_enabled = var.private_dns_enabled
  tags                = merge(var.tags, { Name = var.name })
}