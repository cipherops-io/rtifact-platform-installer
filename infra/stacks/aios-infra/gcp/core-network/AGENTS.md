<!-- DO NOT EDIT — managed by cipherops-io/boilerplates (rtifact-product).
     Change it there, then run: make boilerplate-update -->

# AGENTS.md — core-network (gcp)

## 1. Purpose

Scope, naming, core networking, egress control, hybrid connectivity, jump host.

## 2. Put here / Do NOT put here

**Put here:**

- Resource groups / accounts / projects naming
- VPC/VNet, subnets, NAT, firewall
- Bastion / jump hosts
- VPN / ExpressRoute / Direct Connect


**Do NOT put here:**

- Public ingress — use network-edge
- Managed data services — use data-platform


## 3. Cloud service map (gcp)

| Service | Module folder |
|---------|----------------|
| cloud_nat | `vpc` |
| naming | `naming` |
| vpc | `vpc` |

## 4. File layout

- Use `main.<concern>.tofu` per concern; one `module` block per file; gate with `features.<name>`.

## 5. Upstream inputs and outputs

- Read prior layers via `upstream_state` / `terraform_remote_state`.
- Required outputs: see `docs/layer-contracts.md` for **core-network**.

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