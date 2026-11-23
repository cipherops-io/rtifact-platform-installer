
<!-- BEGIN_TF_DOCS -->


## Providers

| Name | Version |
|------|---------|
| aws | n/a |

## Resources

| Name | Type |
|------|------|
| [aws_route53_record.delegation](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/route53_record) | resource |
| [aws_route53_zone.main](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/route53_zone) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| delegation_ttl | TTL of the NS delegation record in the parent zone (public_dns.delegation_ttl). | `number` | `300` | no |
| name_prefix | n/a | `string` | n/a | yes |
| public_dns | n/a | `any` | n/a | yes |
| tags | n/a | `map(string)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| name_servers | n/a |
| zone_id | n/a |
<!-- END_TF_DOCS -->
