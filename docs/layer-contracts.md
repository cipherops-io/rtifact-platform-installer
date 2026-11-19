# Layer output contracts

Downstream stack layers consume upstream state via `data "terraform_remote_state"` and `var.upstream_state`. Only **output names** are the cross-cloud API.

## core-network

Required outputs (AWS/GCP alias their native IDs to these names):

| Output | Meaning |
|--------|---------|
| `resource_group_name` | Azure RG name; AWS/GCP: scope name / project id string |
| `resource_group_id` | Azure RG id; AWS: account-prefixed scope id; GCP: project id |
| `tags` | Standard tag map (`managed-by = rtifact`) |
| `vnet_id` | VPC / VNet network id |
| `node_subnet_id` | Workload (nodes) subnet |
| `load_balancer_subnet_id` | Ingress/LB subnet |
| `private_endpoint_subnet_id` | Private endpoints / PSC subnet |
| `management_subnet_id` | Bastion / ops subnet |
| `firewall_id` | Azure Firewall (optional) |
| `bastion_host_id` | Managed Azure Bastion (when `bastion.managed`) |

## Other layers

Follow Azure today: wire `core-network` where needed; layer-specific outputs documented in each stack `README.md`.