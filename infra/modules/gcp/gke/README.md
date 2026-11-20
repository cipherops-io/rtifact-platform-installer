
<!-- BEGIN_TF_DOCS -->


## Providers

| Name | Version |
|------|---------|
| google | n/a |

## Resources

| Name | Type |
|------|------|
| [google_container_cluster.main](https://registry.terraform.io/providers/opentofu/google/latest/docs/resources/container_cluster) | resource |
| [google_container_node_pool.system](https://registry.terraform.io/providers/opentofu/google/latest/docs/resources/container_node_pool) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| availability_zones | n/a | `list(string)` | n/a | yes |
| dns_service_ip | n/a | `string` | n/a | yes |
| kubernetes_version | n/a | `string` | n/a | yes |
| location | n/a | `string` | n/a | yes |
| name_prefix | n/a | `string` | n/a | yes |
| network_id | n/a | `string` | n/a | yes |
| node_pools | n/a | <pre>map(object({<br/>    name       = string<br/>    vm_size    = string<br/>    node_count = number<br/>  }))</pre> | n/a | yes |
| node_subnet_id | n/a | `string` | n/a | yes |
| pod_cidr | n/a | `string` | n/a | yes |
| project_id | n/a | `string` | n/a | yes |
| scope_name | n/a | `string` | n/a | yes |
| service_cidr | n/a | `string` | n/a | yes |
| tags | n/a | `map(string)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| cluster_id | n/a |
| cluster_name | n/a |
| oidc_issuer_url | n/a |
<!-- END_TF_DOCS -->
