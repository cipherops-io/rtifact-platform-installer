# Customer runbook

## Config file naming

Use `<customer>-<environment>-<purpose>.yaml` (any path). Example: `tata-digital-ai-os-uat.yaml`. Customer configs stay **outside** the installer repository.

## Prerequisites

- Contributor on the target subscription (or equivalent)
- Identity on the deploy host: **managed identity** (`auth.use_msi: true`) or `ARM_*` credentials
- Quotas for chosen SKUs in the target region
- **Release builds** bundle OpenTofu and an offline provider mirror. Dev builds use `tofu` on `PATH`.

## Customer YAML layout (schema v2)

| Section | Purpose |
|---------|---------|
| `schema_version` | Must be `1` |
| `cloud` | Target cloud (`azure`) |
| `profile` | Workload profile id (e.g. `ai-os-k8s`) |
| `account.subscription_id`, `account.region` | Target subscription and region (`account` replaces the v1 `azure` block) |
| `state.backend`, `state.air_gapped` | `local` or `azurerm`; `air_gapped` defaults to `true` (CLI-only checks; `false` is not implemented yet) |
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
| `redis_cache` | Azure Cache for Redis Premium + PE (mutually exclusive with `azure_managed_redis`) |
| `azure_managed_redis` | Azure Managed Redis + PE |
| `kubernetes_cluster` | Private AKS + node pools in `kubernetes_cluster.node_pools` |
| `azure_bastion` | Jump access: `bastion.managed: true` → Azure Bastion host; `false` → Linux VM on `app-platform` |
| `private_dns_zones` | Private DNS zones for data-layer PEs |
| `azure_firewall` | Azure Firewall + baseline policy (`core-network`) |
| `azure_vpn_gateway` / `azure_express_route_gateway` | VPN or ExpressRoute gateway (one at a time) |
| `private_dns_resolver` | Private DNS resolver + endpoints |
| `public_dns_zone` | Public DNS zones (`network-edge`) |
| `application_gateway` | Application Gateway WAF_v2 placeholder (`network-edge`) |
| `disk_encryption_set` | CMK disk encryption set (`data-platform`) |
| `ai_foundry` | `azurerm_ai_foundry` Hub (`ai-integrations`) |

Setting a feature to `false` creates **no** resources for that block. `features` is a free-form map: keys are validated against `profiles/<profile>/features.yaml` (unknown keys, unsupported features and missing `requires` are rejected).

## Layers

`--stack` selects a layer; `all` runs the five layers after `state-backend` in order and `destroy --stack all` runs them in reverse.

| Layer | Contents |
|-------|----------|
| `state-backend` | Remote state storage in the **same** resource group as the workload (`<name_prefix>-rg`; applied by `bootstrap`, local state) |
| `core-network` | Naming/resource group, VNet and subnets (only layer that creates the resource group) |
| `data-platform` | Key Vault, PostgreSQL, ACR, storage, private DNS (reads `core-network` via remote state) |
| `app-platform` | AKS, bastion (reads `core-network` via remote state) |
| `network-edge` | Application Gateway + public DNS (optional) |
| `ai-integrations` | Communication Services email + AI Foundry hub (optional) |

Stack sources live under `infra/stacks/<profile>/<cloud>/<layer>/`.

`kubernetes_cluster.node_pools` in customer YAML **replaces** profile defaults (no merge of extra pools).

## Deploy

1. Copy `examples/ai-os-centralindia/config.yaml` or an external reference file; set `account.subscription_id` and `workload.name_prefix`.
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
state:
  backend: azurerm
  air_gapped: true
```

Run `installer bootstrap config.yaml` once: it checks (with the `az` CLI) for the state resource group, storage account and `tfstate` container, and applies the `state-backend` layer if they are missing. The `state-backend` layer grants the identity running bootstrap **Storage Blob Data Contributor** on the tfstate account (required because OpenTofu uses `use_azuread_auth` on theazurerm backend). Each layer then keeps its state at `<profile>/<name_prefix>/<layer>.tfstate`.

If plan/apply fails with `AuthorizationPermissionMismatch` on the tfstate account, re-run `rtifact apply config.yaml --stack state-backend --yes` (while signed in as the same principal), or assign **Storage Blob Data Contributor** on that storage account to your user or service principal.

To use a pre-created backend instead, add:

```yaml
backend:
  type: azurerm
  resource_group_name: tfstate-rg
  storage_account_name: tfstatestore
  container_name: tfstate
  key: ai-os-k8s   # prefix; "/<layer>.tfstate" is appended per layer
```

## Outputs after apply

After `app-platform` apply, the CLI writes cluster credentials into the **current working directory**:

| File | Contents |
|------|----------|
| `<name_prefix>-kubeconfig.yaml` | AKS/EKS kubeconfig (private cluster; use bastion or VPN) |
| `<name_prefix>-bastion.pem` | Bastion SSH private key (`chmod 600`) |

Run `rtifact output config.yaml` anytime to print bastion IP, cluster name, Key Vault URI, and paths to those local files. Layer JSON snapshots live under `~/.rtifact/<name_prefix>/`.

Use `rtifact output config.yaml --tofu --stack app-platform` for raw OpenTofu `-json` output.

## Backup

With local state, state files live under `~/.rtifact/<name_prefix>/state/`; back them up after every successful apply.
