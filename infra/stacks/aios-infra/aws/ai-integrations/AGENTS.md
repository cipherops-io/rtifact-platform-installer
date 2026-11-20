<!-- DO NOT EDIT — managed by cipherops-io/boilerplates (rtifact-product).
     Change it there, then run: make boilerplate-update -->

# AGENTS.md — ai-integrations (aws)

## 1. Purpose

Model platform access and transactional email / communications.

## 2. Put here / Do NOT put here

**Put here:**

- AI hub / Foundry / Bedrock / Vertex endpoints
- Email and communication services


**Do NOT put here:**

- General data stores — use data-platform


## 3. Cloud service map (aws)

| Service | Module folder |
|---------|----------------|
| bedrock | `ai-integrations` |

## 4. File layout

- Use `main.<concern>.tofu` per concern; one `module` block per file; gate with `features.<name>`.

## 5. Upstream inputs and outputs

- Read prior layers via `upstream_state` / `terraform_remote_state`.
- Required outputs: see `docs/layer-contracts.md` for **ai-integrations**.

## 6. Add-a-service checklist

1. Create `infra/modules/aws/<name>/`
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