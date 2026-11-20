
<!-- BEGIN_TF_DOCS -->


## Providers

| Name | Version |
|------|---------|
| google | n/a |

## Resources

| Name | Type |
|------|------|
| [google_compute_network.main](https://registry.terraform.io/providers/opentofu/google/latest/docs/resources/compute_network) | resource |
| [google_compute_router.main](https://registry.terraform.io/providers/opentofu/google/latest/docs/resources/compute_router) | resource |
| [google_compute_router_nat.main](https://registry.terraform.io/providers/opentofu/google/latest/docs/resources/compute_router_nat) | resource |
| [google_compute_subnetwork.load_balancer](https://registry.terraform.io/providers/opentofu/google/latest/docs/resources/compute_subnetwork) | resource |
| [google_compute_subnetwork.management](https://registry.terraform.io/providers/opentofu/google/latest/docs/resources/compute_subnetwork) | resource |
| [google_compute_subnetwork.node](https://registry.terraform.io/providers/opentofu/google/latest/docs/resources/compute_subnetwork) | resource |
| [google_compute_subnetwork.private_endpoint](https://registry.terraform.io/providers/opentofu/google/latest/docs/resources/compute_subnetwork) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| load_balancer_subnet_cidr | n/a | `string` | n/a | yes |
| location | n/a | `string` | n/a | yes |
| management_subnet_cidr | n/a | `string` | n/a | yes |
| name_prefix | n/a | `string` | n/a | yes |
| node_subnet_cidr | n/a | `string` | n/a | yes |
| private_endpoint_subnet_cidr | n/a | `string` | n/a | yes |
| project_id | n/a | `string` | n/a | yes |
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
