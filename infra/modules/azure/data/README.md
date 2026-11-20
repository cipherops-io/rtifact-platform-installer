# data

PostgreSQL, Redis, ADLS corpus storage, Foundry non-HNS storage, ACR, private DNS, and private endpoints.
<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
|------|---------|
| terraform | >= 1.6.0 |
| random | >= 3.0 |

## Providers

| Name | Version |
|------|---------|
| azurerm | n/a |
| random | >= 3.0 |

## Resources

| Name | Type |
|------|------|
| [azurerm_container_registry.main](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/container_registry) | resource |
| [azurerm_managed_redis.main](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/managed_redis) | resource |
| [azurerm_postgresql_flexible_server.main](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/postgresql_flexible_server) | resource |
| [azurerm_postgresql_flexible_server.read_replica](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/postgresql_flexible_server) | resource |
| [azurerm_postgresql_flexible_server_configuration.extensions](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/postgresql_flexible_server_configuration) | resource |
| [azurerm_private_dns_zone.acr](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/private_dns_zone) | resource |
| [azurerm_private_dns_zone.blob](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/private_dns_zone) | resource |
| [azurerm_private_dns_zone.file](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/private_dns_zone) | resource |
| [azurerm_private_dns_zone.postgres](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/private_dns_zone) | resource |
| [azurerm_private_dns_zone.redis](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/private_dns_zone) | resource |
| [azurerm_private_dns_zone_virtual_network_link.acr](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/private_dns_zone_virtual_network_link) | resource |
| [azurerm_private_dns_zone_virtual_network_link.blob](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/private_dns_zone_virtual_network_link) | resource |
| [azurerm_private_dns_zone_virtual_network_link.file](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/private_dns_zone_virtual_network_link) | resource |
| [azurerm_private_dns_zone_virtual_network_link.postgres](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/private_dns_zone_virtual_network_link) | resource |
| [azurerm_private_dns_zone_virtual_network_link.redis](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/private_dns_zone_virtual_network_link) | resource |
| [azurerm_private_endpoint.acr](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/private_endpoint) | resource |
| [azurerm_private_endpoint.blob](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/private_endpoint) | resource |
| [azurerm_private_endpoint.files](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/private_endpoint) | resource |
| [azurerm_private_endpoint.foundry_blob](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/private_endpoint) | resource |
| [azurerm_private_endpoint.managed_redis](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/private_endpoint) | resource |
| [azurerm_private_endpoint.postgres](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/private_endpoint) | resource |
| [azurerm_private_endpoint.redis](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/private_endpoint) | resource |
| [azurerm_redis_cache.main](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/redis_cache) | resource |
| [azurerm_storage_account.files](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/storage_account) | resource |
| [azurerm_storage_account.foundry](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/storage_account) | resource |
| [azurerm_storage_account.main](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/storage_account) | resource |
| [azurerm_storage_container.backup](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/storage_container) | resource |
| [azurerm_storage_container.corpus](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/storage_container) | resource |
| [azurerm_storage_share.aios](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/storage_share) | resource |
| [random_password.pg](https://registry.terraform.io/providers/opentofu/random/latest/docs/resources/password) | resource |
| [random_string.files_suffix](https://registry.terraform.io/providers/opentofu/random/latest/docs/resources/string) | resource |
| [random_string.foundry_suffix](https://registry.terraform.io/providers/opentofu/random/latest/docs/resources/string) | resource |
| [random_string.suffix](https://registry.terraform.io/providers/opentofu/random/latest/docs/resources/string) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| acr_sku | n/a | `string` | `"Premium"` | no |
| azure_files_quota_gb | n/a | `number` | `1024` | no |
| azure_files_share_name | n/a | `string` | `"aios-shared"` | no |
| azure_files_zone_redundant | n/a | `bool` | `true` | no |
| dns_resource_group_name | n/a | `string` | n/a | yes |
| features | n/a | <pre>object({<br/>    postgresql_flexible_server = bool<br/>    container_registry         = bool<br/>    storage_account            = bool<br/>    redis_cache                = bool<br/>    azure_managed_redis        = optional(bool, false)<br/>    private_dns_zones          = bool<br/>    azure_files_share          = bool<br/>    disk_encryption_set        = optional(bool, false)<br/>    ai_foundry                 = optional(bool, false)<br/>  })</pre> | n/a | yes |
| foundry_storage | Non-HNS storage for Azure AI Foundry; set from customer ai_foundry.storage YAML. | <pre>object({<br/>    account_tier             = string<br/>    account_replication_type = string<br/>    public_network_access    = string<br/>  })</pre> | n/a | yes |
| location | n/a | `string` | n/a | yes |
| managed_redis_sku_name | n/a | `string` | `"Balanced_B3"` | no |
| name_prefix | n/a | `string` | n/a | yes |
| postgresql_backup_retention_days | n/a | `number` | `7` | no |
| postgresql_ha_enabled | n/a | `bool` | `false` | no |
| postgresql_ha_mode | n/a | `string` | `"ZoneRedundant"` | no |
| postgresql_read_replica_enabled | n/a | `bool` | `false` | no |
| postgresql_sku_name | n/a | `string` | `"GP_Standard_D8ds_v5"` | no |
| postgresql_storage_iops | n/a | `number` | `null` | no |
| postgresql_storage_mb | n/a | `number` | `1048576` | no |
| postgresql_version | n/a | `string` | `"16"` | no |
| private_endpoint_subnet_id | n/a | `string` | n/a | yes |
| redis_capacity | n/a | `number` | `1` | no |
| redis_family | n/a | `string` | `"P"` | no |
| redis_minimum_tls_version | n/a | `string` | `"1.2"` | no |
| redis_sku_tier | n/a | `string` | `"Premium"` | no |
| resource_group_name | n/a | `string` | n/a | yes |
| storage_backup_gb | n/a | `number` | `250` | no |
| storage_corpus_gb | n/a | `number` | `250` | no |
| tags | n/a | `map(string)` | `{}` | no |
| vnet_id | n/a | `string` | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| acr_id | n/a |
| ai_foundry_storage_account_id | n/a |
| azure_files_share_name | n/a |
| azure_files_storage_account_id | n/a |
| managed_redis_id | n/a |
| postgresql_id | n/a |
| private_dns_zone_ids | n/a |
| redis_id | n/a |
| storage_account_id | n/a |
<!-- END_TF_DOCS -->
