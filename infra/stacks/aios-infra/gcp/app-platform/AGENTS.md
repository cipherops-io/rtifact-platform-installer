<!-- DO NOT EDIT — managed by cipherops-io/boilerplates (rtifact-product).
     Change it there, then run: make boilerplate-update -->

# AGENTS.md — app-platform (gcp)

## 1. Purpose

Kubernetes clusters, node pools, workload identity, compute VMs.

## 2. Put here / Do NOT put here

**Put here:**

- AKS/EKS/GKE and node pools
- VM scale sets / EC2 / GCE when not on K8s


**Do NOT put here:**

- Databases — use data-platform


## 3. Cloud service map (gcp)

| Service | Module folder |
|---------|----------------|
| compute | `compute` |
| gke | `gke` |

## 4. File layout

- Use `main.<concern>.tofu` per concern; one `module` block per file; gate with `features.<name>`.

## 5. Upstream inputs and outputs

- Read prior layers via `upstream_state` / `terraform_remote_state`.
- Required outputs: see `docs/layer-contracts.md` for **app-platform**.

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