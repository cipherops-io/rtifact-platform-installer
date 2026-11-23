
<!-- BEGIN_TF_DOCS -->


## Providers

| Name | Version |
|------|---------|
| aws | n/a |

## Resources

| Name | Type |
|------|------|
| [aws_vpc_endpoint.main](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/vpc_endpoint) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| location | n/a | `string` | n/a | yes |
| name | n/a | `string` | n/a | yes |
| private_dns_enabled | Associate the AWS-managed private hosted zone with the VPC so service hostnames resolve to the endpoint. | `bool` | `true` | no |
| scope_name | n/a | `string` | n/a | yes |
| service_name | n/a | `string` | n/a | yes |
| subnet_id | n/a | `string` | n/a | yes |
| tags | n/a | `map(string)` | `{}` | no |
| vpc_id | n/a | `string` | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| endpoint_id | n/a |
<!-- END_TF_DOCS -->
