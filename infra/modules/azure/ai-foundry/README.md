# ai-foundry

Provisions a classic Microsoft Foundry Hub via `azurerm_ai_foundry`. Uses dedicated non-HNS storage and access-policy Key Vault outputs from `data-platform`. Hub MSI Key Vault access is created by Azure during workspace provisioning (not managed here). Model deployments remain manual.
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
| [azurerm_ai_foundry.main](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/ai_foundry) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| container_registry_id | n/a | `string` | `null` | no |
| key_vault_id | n/a | `string` | n/a | yes |
| location | n/a | `string` | n/a | yes |
| name_prefix | n/a | `string` | n/a | yes |
| public_network_access | n/a | `string` | `"Disabled"` | no |
| resource_group_name | n/a | `string` | n/a | yes |
| storage_account_id | n/a | `string` | n/a | yes |
| tags | n/a | `map(string)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| ai_foundry_id | n/a |
| workspace_id | n/a |
<!-- END_TF_DOCS -->
