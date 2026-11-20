
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
| [aws_route_table.private](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/route_table) | resource |
| [aws_route_table_association.management](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/route_table_association) | resource |
| [aws_route_table_association.node](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/route_table_association) | resource |
| [aws_route_table_association.private_endpoint](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/route_table_association) | resource |
| [aws_subnet.load_balancer](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/subnet) | resource |
| [aws_subnet.management](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/subnet) | resource |
| [aws_subnet.node](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/subnet) | resource |
| [aws_subnet.private_endpoint](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/subnet) | resource |
| [aws_vpc.main](https://registry.terraform.io/providers/opentofu/aws/latest/docs/resources/vpc) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| load_balancer_subnet_cidr | n/a | `string` | n/a | yes |
| location | n/a | `string` | n/a | yes |
| management_subnet_cidr | n/a | `string` | n/a | yes |
| name_prefix | n/a | `string` | n/a | yes |
| node_subnet_cidr | n/a | `string` | n/a | yes |
| private_endpoint_subnet_cidr | n/a | `string` | n/a | yes |
| scope_name | n/a | `string` | n/a | yes |
| tags | n/a | `map(string)` | `{}` | no |
| vnet_cidr | n/a | `string` | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| load_balancer_subnet_id | n/a |
| management_subnet_id | n/a |
| nat_gateway_id | n/a |
| node_subnet_id | n/a |
| private_endpoint_subnet_id | n/a |
| vnet_id | n/a |
| vnet_name | n/a |
<!-- END_TF_DOCS -->
