
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
| [azurerm_disk_encryption_set.main](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/disk_encryption_set) | resource |
| [azurerm_key_vault.foundry](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/key_vault) | resource |
| [azurerm_key_vault.main](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/key_vault) | resource |
| [azurerm_key_vault_access_policy.foundry_deployer](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/key_vault_access_policy) | resource |
| [azurerm_key_vault_key.des](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/key_vault_key) | resource |
| [azurerm_role_assignment.deployer_crypto_officer](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/role_assignment) | resource |
| [azurerm_role_assignment.des_crypto_user](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/role_assignment) | resource |
| [azurerm_role_assignment.foundry_deployer_contributor](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/role_assignment) | resource |
| [azurerm_user_assigned_identity.des](https://registry.terraform.io/providers/opentofu/azurerm/latest/docs/resources/user_assigned_identity) | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| des_key | Disk encryption set key (disk_encryption_set.key in customer YAML). | <pre>object({<br/>    key_type                  = optional(string, "RSA")<br/>    key_size                  = optional(number, 3072)<br/>    auto_key_rotation_enabled = optional(bool, false)<br/>  })</pre> | `{}` | no |
| features | n/a | <pre>object({<br/>    disk_encryption_set = optional(bool, false)<br/>    ai_foundry          = optional(bool, false)<br/>  })</pre> | <pre>{<br/>  "ai_foundry": false,<br/>  "disk_encryption_set": false<br/>}</pre> | no |
| foundry_key_vault | Foundry-dedicated Key Vault (access-policy model); set from customer key_vault.foundry / ai_foundry.key_vault YAML. | <pre>object({<br/>    sku                              = string<br/>    soft_delete_retention_days       = number<br/>    purge_protection_enabled         = bool<br/>    grant_deployer_contributor       = bool<br/>    deployer_key_permissions         = list(string)<br/>    deployer_secret_permissions      = list(string)<br/>    deployer_certificate_permissions = list(string)<br/>  })</pre> | n/a | yes |
| key_vault | Workload Key Vault settings (key_vault in customer YAML). | <pre>object({<br/>    soft_delete_retention_days = optional(number, 7)<br/>    purge_protection_enabled   = optional(bool, true)<br/>    rbac_authorization_enabled = optional(bool, true)<br/>  })</pre> | `{}` | no |
| location | n/a | `string` | n/a | yes |
| name_prefix | n/a | `string` | n/a | yes |
| resource_group_name | n/a | `string` | n/a | yes |
| sku | Key Vault SKU (standard or premium). | `string` | `"premium"` | no |
| tags | n/a | `map(string)` | `{}` | no |
| tenant_id | n/a | `string` | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| ai_foundry_key_vault_id | n/a |
| ai_foundry_key_vault_uri | n/a |
| disk_encryption_set_id | n/a |
| key_vault_id | n/a |
| key_vault_uri | n/a |
<!-- END_TF_DOCS -->
