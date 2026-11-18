# ai-os-aks stack

OpenTofu composition root. Runtime: `tofu` only (via `installer` CLI).

- `providers.tofu` — provider pins from [OpenTofu Registry](https://search.opentofu.org/providers) (`opentofu/azurerm`, `opentofu/random`, `opentofu/tls` only)
- `.terraform.lock.hcl` — committed provider lock (OpenTofu registry)
- Customer values arrive as `default.auto.tofuvars.json` (generated at apply time from `config.yaml`)

```bash
tofu init -backend=false
tofu validate
```