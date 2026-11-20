<!-- DO NOT EDIT — managed by cipherops-io/boilerplates (rtifact-product).
     Change it there, then run: make boilerplate-update -->

# AGENTS.md — network-edge (gcp)

## 1. Purpose

Public ingress, edge security, public DNS, edge TLS termination.

## 2. Put here / Do NOT put here

**Put here:**

- Application gateways / ALB+WAF / global LB
- Public DNS zones and records


**Do NOT put here:**

- Core VNet routing — use core-network


## 3. Cloud service map (gcp)

| Service | Module folder |
|---------|----------------|
| cloud_dns | `dns-public` |
| external_lb | `network-edge` |

## 4. File layout

- Use `main.<concern>.tofu` per concern; one `module` block per file; gate with `features.<name>`.

## 5. Upstream inputs and outputs

- Read prior layers via `upstream_state` / `terraform_remote_state`.
- Required outputs: see `docs/layer-contracts.md` for **network-edge**.

## 6. Add-a-service checklist

1. Create `infra/modules/gcp/<name>/`
2. Add module to `Modules` in `boilerplate-vars.yml`
3. Add `features.<name>` to profile + schema
4. Wire in this stack layer
5. `make tofu-pre-commit-fast`

## 7. Validate this layer

```bash
make tofu-validate
```

## 8. Installer-specific notes

See `AGENTS.local.md` in the repo root.