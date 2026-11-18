# Customer runbook

## Config file naming

Use `<customer>-<environment>-<purpose>.yaml` (any path). Example: `tata-digital-ai-os-uat.yaml`. Customer configs stay **outside** the installer repository.

## Prerequisites

- Contributor on the target subscription (or equivalent)
- Identity on the deploy host: **managed identity** (`auth.use_msi: true`) or `ARM_*` credentials
- Quotas for chosen SKUs in the target region
- **Release builds** bundle OpenTofu and an offline provider mirror. Dev builds use `tofu` on `PATH`.

## Customer YAML layout (v1)

| Section | Purpose |
|---------|---------|
| `profile` | Workload profile id (e.g. `ai-os-aks`) |
| `azure.subscription_id`, `azure.region` | Target subscription and region |
| `workload.name_prefix` | Resource naming prefix |
| `features.*` | Boolean gates for each deployable block (maps 1:1 to OpenTofu `var.features`) |
| `network`, `kubernetes_cluster`, `key_vault`, … | Sizing overrides (merged with profile defaults) |

### Feature flags

| Flag | Deploys |
|------|---------|
| `resource_group` | Workload resource group |
| `virtual_network` | VNet, subnets, NAT |
| `key_vault` | Key Vault |
| `key_vault_private_endpoint` | Key Vault private endpoint + DNS |
| `postgresql_flexible_server` | PostgreSQL Flexible Server + PE |
| `container_registry` | ACR + PE |
| `storage_account` | Premium ADLS storage + PE |
| `redis_cache` | **Unsupported** in IaC (keep `false`) |
| `kubernetes_cluster` | Private AKS + node pools in `kubernetes_cluster.node_pools` |
| `azure_bastion` | Bastion host |
| `private_dns_zones` | Private DNS zones for data-layer PEs |

Setting a feature to `false` creates **no** resources for that block.

`kubernetes_cluster.node_pools` in customer YAML **replaces** profile defaults (no merge of extra pools).

## Deploy

1. Copy `examples/ai-os-centralindia/config.yaml` or an external reference file; set `azure.subscription_id` and `workload.name_prefix`.
2. Validate:

   ```bash
   ./installer validate config.yaml
   ```

3. Plan:

   ```bash
   ./installer plan config.yaml
   ```

4. Apply:

   ```bash
   ./installer apply config.yaml
   ```

5. Complete the printed post-apply checklist (Atlas, email, firewall allow-lists, DNS).

## Remote state (recommended for production)

```yaml
backend:
  type: azurerm
  resource_group_name: tfstate-rg
  storage_account_name: tfstatestore
  container_name: tfstate
  key: ai-os-aks.tfstate
```

Pre-create the storage account and container before first `apply`.

## Backup

When using local state, back up `tofu.tfstate` after every successful apply.