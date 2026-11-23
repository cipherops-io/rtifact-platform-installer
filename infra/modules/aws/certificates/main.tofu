locals {
  cfg = var.certificates
}

resource "aws_acm_certificate" "regional" {

  domain_name               = local.cfg.names[0]
  subject_alternative_names = slice(local.cfg.names, 1, length(local.cfg.names))
  validation_method         = "DNS"
  key_algorithm             = var.key_algorithm
  tags                      = var.tags
}

resource "aws_route53_record" "validation" {
  for_each = {
    for dvo in aws_acm_certificate.regional.domain_validation_options : dvo.domain_name => {
      name   = dvo.resource_record_name
      record = dvo.resource_record_value
      type   = dvo.resource_record_type
    }
  }

  allow_overwrite = true
  name            = each.value.name
  records         = [each.value.record]
  ttl             = var.validation_record_ttl
  type            = each.value.type
  zone_id         = var.zone_id
}

resource "aws_acm_certificate_validation" "regional" {
  certificate_arn         = aws_acm_certificate.regional.arn
  validation_record_fqdns = [for r in aws_route53_record.validation : r.fqdn]
}

resource "aws_acm_certificate" "cloudfront" {
  provider = aws.us_east_1
  count    = try(local.cfg.cloudfront, false) ? 1 : 0

  domain_name               = local.cfg.names[0]
  subject_alternative_names = slice(local.cfg.names, 1, length(local.cfg.names))
  validation_method         = "DNS"
  key_algorithm             = var.key_algorithm
  tags                      = var.tags
}

resource "aws_route53_record" "validation_cloudfront" {
  for_each = try(local.cfg.cloudfront, false) ? {
    for dvo in aws_acm_certificate.cloudfront[0].domain_validation_options : dvo.domain_name => {
      name   = dvo.resource_record_name
      record = dvo.resource_record_value
      type   = dvo.resource_record_type
    }
  } : {}

  allow_overwrite = true
  name            = each.value.name
  records         = [each.value.record]
  ttl             = var.validation_record_ttl
  type            = each.value.type
  zone_id         = var.zone_id
}

resource "aws_acm_certificate_validation" "cloudfront" {
  provider = aws.us_east_1
  count    = try(local.cfg.cloudfront, false) ? 1 : 0

  certificate_arn         = aws_acm_certificate.cloudfront[0].arn
  validation_record_fqdns = [for r in aws_route53_record.validation_cloudfront : r.fqdn]
}