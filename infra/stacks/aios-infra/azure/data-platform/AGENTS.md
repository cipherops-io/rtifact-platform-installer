<!-- DO NOT EDIT — managed by cipherops-io/boilerplates (rtifact-product).
     Change it there, then run: make boilerplate-update -->

# AGENTS.md — data-platform (azure)

## 1. Purpose

Stateful data, secrets, registries, private endpoints and private DNS.

## 2. Put here / Do NOT put here

**Put here:**

- Databases, caches, object storage, file shares
- Key vault / KMS, CMK, RBAC for data plane
- Private endpoints and DNS for PaaS


**Do NOT put here:**

- Kubernetes — use app-platform
- Edge TLS / WAF — use network-edge


## 3. Cloud service map (azure)

| Service | Module folder |
|---------|----------------|
| data | `data` |
| private-endpoint | `private-endpoint` |
| security | `security` |

## 4. File layout

- Use `main.<concern>.tofu` per concern; one `module` block per file; gate with `features.<name>`.

## 5. Upstream inputs and outputs

- Read prior layers via `upstream_state` / `terraform_remote_state`.
- Required outputs: see `docs/layer-contracts.md` for **data-platform**.

## 6. Add-a-service checklist

1. Create `infra/modules/azure/<name>/`
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