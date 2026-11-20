# Layer output contracts

Downstream stack layers consume upstream state via `data "terraform_remote_state"` and `var.upstream_state`. Only **output names** are the cross-cloud API. The **all three clouds** column is enforced by `tests/config/layer_contracts_test.go`.

## state-backend {#state-backend}

| Output | All three clouds | Azure only | AWS / GCP only |
|--------|------------------|------------|----------------|
| `resource_group_name` | yes | | |
| `storage_account_name` | | yes | AWS `s3_bucket_name`, GCP `gcs_bucket_name` |
| `dynamodb_table_name` | | | AWS |
| `s3_bucket_name` | | | AWS |

## core-network {#core-network}

| Output | All three clouds | Azure only | AWS / GCP only |
|--------|------------------|------------|----------------|
| `resource_group_name` | yes | | |
| `resource_group_id` | yes | | |
| `tags` | yes | | |
| `vnet_id` | yes | | |
| `node_subnet_id` | yes | | |
| `load_balancer_subnet_id` | yes | | |
| `private_endpoint_subnet_id` | yes | | |
| `management_subnet_id` | yes | | |
| `firewall_id` | | yes | |
| `vpn_gateway_id` | | yes | |
| `express_route_gateway_id` | | yes | |
| `bastion_host_id` | | yes | |
| `private_dns_resolver_id` | | yes | |

## data-platform {#data-platform}

| Output | All three clouds | Azure only | AWS / GCP only |
|--------|------------------|------------|----------------|
| `key_vault_id` | yes | | |
| `key_vault_uri` | yes | | |
| `postgresql_id` | yes | | |
| `storage_account_id` | yes | | |
| `acr_id` | yes | | |
| `ai_foundry_key_vault_id` | | yes | |
| `ai_foundry_key_vault_uri` | | yes | |
| `ai_foundry_storage_account_id` | | yes | |
| `redis_id` | | yes | |
| `managed_redis_id` | | yes | |
| `disk_encryption_set_id` | | yes | |
| `azure_files_share_name` | | yes | |

## app-platform {#app-platform}

| Output | All three clouds | Azure only | AWS / GCP only |
|--------|------------------|------------|----------------|
| `cluster_name` | yes | | |
| `aks_name` | yes | | |
| `oidc_issuer_url` | yes | | |
| `bastion_private_ip` | yes | | |
| `bastion_ssh_private_key_pem` (sensitive) | yes | | |
| `kubeconfig_raw` (sensitive) | | yes | |
| `kube_admin_config_raw` (sensitive) | | yes | |
| `cluster_endpoint` | | | AWS |
| `cluster_certificate_authority_data` | | | AWS |

## network-edge {#network-edge}

| Output | All three clouds | Azure only | AWS / GCP only |
|--------|------------------|------------|----------------|
| `resource_group_name` | | | AWS / GCP stubs pass through |
| `application_gateway_id` | | yes | |
| `public_dns_zone_ids` | | yes | |

## ai-integrations {#ai-integrations}

| Output | All three clouds | Azure only | AWS / GCP only |
|--------|------------------|------------|----------------|
| `resource_group_name` | | | AWS / GCP stubs |
| `communication_service_id` | | yes | |
| `email_communication_service_id` | | yes | |
| `ai_foundry_id` | | yes | |