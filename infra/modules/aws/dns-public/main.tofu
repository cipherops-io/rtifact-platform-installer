locals {
  cfg = var.public_dns
}

resource "aws_route53_zone" "main" {
  name = local.cfg.zone_name
  tags = merge(var.tags, { Name = local.cfg.zone_name })
}

resource "aws_route53_record" "delegation" {
  count = try(local.cfg.parent_zone_id, "") != "" ? 1 : 0

  zone_id = local.cfg.parent_zone_id
  name    = trimsuffix(local.cfg.zone_name, ".")
  type    = "NS"
  ttl     = 300
  records = aws_route53_zone.main.name_servers
}