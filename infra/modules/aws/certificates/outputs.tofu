output "regional_certificate_arn" {
  value = aws_acm_certificate.regional.arn
}

output "cloudfront_certificate_arn" {
  value = try(aws_acm_certificate.cloudfront[0].arn, null)
}