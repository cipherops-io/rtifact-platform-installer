
<!-- BEGIN_TF_DOCS -->


## Providers

| Name | Version |
|------|---------|
| aws | n/a |

## Resources

| Name | Type |
|------|------|
| [aws_eip.nat](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/eip) | resource |
| [aws_internet_gateway.main](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/internet_gateway) | resource |
| [aws_nat_gateway.main](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/nat_gateway) | resource |
| [aws_route.private_nat](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/route) | resource |
| [aws_route_table.isolated](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/route_table) | resource |
| [aws_route_table.private](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/route_table) | resource |
| [aws_route_table.public](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/route_table) | resource |
| [aws_route_table_association.tier](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/route_table_association) | resource |
| [aws_subnet.tier](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/subnet) | resource |
| [aws_vpc.main](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/vpc) | resource |
| [aws_vpc_endpoint.gateway](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/vpc_endpoint) | resource |
| [aws_vpc_endpoint_route_table_association.gateway](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/vpc_endpoint_route_table_association) | resource |
| [aws_vpc_ipv4_cidr_block_association.secondary](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/vpc_ipv4_cidr_block_association) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| cluster_name | EKS cluster name for kubernetes.io/cluster/* subnet tags | `string` | n/a | yes |
| location | n/a | `string` | n/a | yes |
| name_prefix | n/a | `string` | n/a | yes |
| network | n/a | <pre>object({<br/>    vpc_cidr             = string<br/>    secondary_cidrs      = optional(list(string), [])<br/>    availability_zones   = list(string)<br/>    nat_gateway          = optional(string, "single")<br/>    gateway_endpoints    = optional(list(string), [])<br/>    flow_logs            = optional(bool, false)<br/>    enable_dns_support   = optional(bool, true)<br/>    enable_dns_hostnames = optional(bool, true)<br/>    subnets = map(object({<br/>      cidrs               = list(string)<br/>      route               = string<br/>      public_ip_on_launch = optional(bool, false)<br/>      roles               = optional(list(string), [])<br/>    }))<br/>  })</pre> | n/a | yes |
| tags | n/a | `map(string)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| availability_zones | n/a |
| internet_gateway_id | n/a |
| nat_gateway_public_ips | n/a |
| route_table_ids | n/a |
| subnet_ids_by_tier | n/a |
| vpc_cidrs | n/a |
| vpc_id | n/a |
<!-- END_TF_DOCS -->
