
<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
|------|---------|
| terraform | >= 1.6.0 |

## Providers

| Name | Version |
|------|---------|
| azurerm | n/a |

## Resources

| Name | Type |
|------|------|
| [azurerm_kubernetes_cluster.main](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/kubernetes_cluster) | resource |
| [azurerm_kubernetes_cluster_node_pool.user](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/kubernetes_cluster_node_pool) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| availability_zones | AKS node pool AZs (centralindia supports 1 and 2 only). | `list(string)` | <pre>[<br/>  "1",<br/>  "2",<br/>  "3"<br/>]</pre> | no |
| dns_service_ip | n/a | `string` | n/a | yes |
| kubernetes_version | n/a | `string` | n/a | yes |
| location | n/a | `string` | n/a | yes |
| name_prefix | n/a | `string` | n/a | yes |
| node_pools | User node pools keyed by role. | <pre>map(object({<br/>    name       = string<br/>    vm_size    = string<br/>    node_count = number<br/>  }))</pre> | n/a | yes |
| node_subnet_id | n/a | `string` | n/a | yes |
| os_disk_type | AKS node OS disk type: Managed or Ephemeral (not Azure storage SKU names). | `string` | `"Managed"` | no |
| pod_cidr | n/a | `string` | n/a | yes |
| resource_group_name | n/a | `string` | n/a | yes |
| service_cidr | n/a | `string` | n/a | yes |
| tags | n/a | `map(string)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| cluster_id | n/a |
| cluster_name | n/a |
| kube_admin_config_raw | n/a |
| kubeconfig_raw | n/a |
| oidc_issuer_url | n/a |
<!-- END_TF_DOCS -->
