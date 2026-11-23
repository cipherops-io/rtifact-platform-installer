
<!-- BEGIN_TF_DOCS -->


## Providers

| Name | Version |
|------|---------|
| aws | n/a |
| aws.us_east_1 | n/a |

## Resources

| Name | Type |
|------|------|
| [aws_acm_certificate.cloudfront](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/acm_certificate) | resource |
| [aws_acm_certificate.regional](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/acm_certificate) | resource |
| [aws_acm_certificate_validation.cloudfront](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/acm_certificate_validation) | resource |
| [aws_acm_certificate_validation.regional](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/acm_certificate_validation) | resource |
| [aws_route53_record.validation](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/route53_record) | resource |
| [aws_route53_record.validation_cloudfront](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/route53_record) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| certificates | n/a | `any` | n/a | yes |
| key_algorithm | ACM key algorithm (certificates.key_algorithm). | `string` | `"RSA_2048"` | no |
| location | n/a | `string` | n/a | yes |
| name_prefix | n/a | `string` | n/a | yes |
| tags | n/a | `map(string)` | `{}` | no |
| validation_record_ttl | TTL of the ACM DNS validation records (certificates.validation_record_ttl). | `number` | `60` | no |
| zone_id | n/a | `string` | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| cloudfront_certificate_arn | n/a |
| regional_certificate_arn | n/a |
<!-- END_TF_DOCS -->
