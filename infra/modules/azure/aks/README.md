
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
| api_server | API server exposure (kubernetes_cluster.api_server). Changing private_cluster_enabled or private_dns_zone_id recreates the cluster. | <pre>object({<br/>    private_cluster_enabled = optional(bool, true)<br/>    private_dns_zone_id     = optional(string, "System")<br/>    authorized_ip_ranges    = optional(list(string), [])<br/>  })</pre> | `{}` | no |
| availability_zones | AKS node pool AZs (centralindia supports 1 and 2 only). | `list(string)` | <pre>[<br/>  "1",<br/>  "2",<br/>  "3"<br/>]</pre> | no |
| cluster | AKS control plane settings (kubernetes_cluster in customer YAML). | <pre>object({<br/>    sku_tier                  = optional(string, "Standard")<br/>    oidc_issuer_enabled       = optional(bool, true)<br/>    workload_identity_enabled = optional(bool, true)<br/>    node_provisioning_mode    = optional(string, "Manual")<br/>  })</pre> | `{}` | no |
| dns_service_ip | n/a | `string` | n/a | yes |
| kubernetes_version | n/a | `string` | n/a | yes |
| location | n/a | `string` | n/a | yes |
| name_prefix | n/a | `string` | n/a | yes |
| network_profile | AKS network profile (kubernetes_cluster.network_profile). | <pre>object({<br/>    network_plugin      = optional(string, "azure")<br/>    network_plugin_mode = optional(string, "overlay")<br/>    outbound_type       = optional(string, "userAssignedNATGateway")<br/>  })</pre> | `{}` | no |
| node_pools | Node pools keyed by role. The `system` key overrides system_pool; the rest become user pools. | <pre>map(object({<br/>    name            = optional(string)<br/>    vm_size         = string<br/>    node_count      = number<br/>    os_disk_size_gb = optional(number, 128)<br/>    mode            = optional(string, "User")<br/>  }))</pre> | n/a | yes |
| node_subnet_id | n/a | `string` | n/a | yes |
| os_disk_type | AKS node OS disk type: Managed or Ephemeral (not Azure storage SKU names). | `string` | `"Managed"` | no |
| pod_cidr | n/a | `string` | n/a | yes |
| resource_group_name | n/a | `string` | n/a | yes |
| service_cidr | n/a | `string` | n/a | yes |
| system_pool | Default (system) node pool used when node_pools has no `system` entry (kubernetes_cluster.system_pool). | <pre>object({<br/>    name            = optional(string, "system")<br/>    vm_size         = optional(string, "Standard_D4s_v5")<br/>    node_count      = optional(number, 3)<br/>    os_disk_size_gb = optional(number, 128)<br/>  })</pre> | `{}` | no |
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
