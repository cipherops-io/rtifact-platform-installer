# Changelog

## [0.5.0] - 2026-11-23

Aligned with **boilerplates [v3.3.0](https://github.com/cipherops-io/boilerplates/releases/tag/v3.3.0)**.

### Deploy engine

- `rtifact deploy plan|apply|status|rollback|destroy` with OCI charts, identity guard, inventory in `rtifact-system`, and lease locking.
- Level-2 CI: `make deploy-e2e` (kind + Zot). See `tests/deploy/README.md` for scenario coverage.
- AWS EKS **pod_identities** (preview): module, schema, `pod_identity_role_arns` output; `examples/aws/aws-sbx2.yaml` updated.

## [0.4.1] - 2026-11-22

- `boilerplate-source.yaml` records the canonical GitHub `template_url` so `make boilerplate-verify` matches CI.
- `prepareLayerMerged` takes `cloud.Cloud` directly (no `Lifecycle` type assertions in the runner).
- `.gitignore`: `.cursor/` and generated `internal/packaged/examples/`.

## [0.4.0] - 2026-11-22

Aligned with **boilerplates [v3.2.1](https://github.com/cipherops-io/boilerplates/releases/tag/v3.2.1)** (`a8c2b61` in `boilerplate-source.yaml`).

No customer config contract changes (`schema_version: 1`). Generated tfvars, backends, and layer order unchanged (golden tests).

### Structure (this release)

- One `cloud.Cloud` interface; `internal/cloudconfig` and runtime hook lookups removed.
- Azure resource-group handling behind `PrepareLayerVars`; shared `internal/engine` plan/confirm/act/record path.
- Profiles are data: checklists, aliases, and packaging read from `profiles/<id>/`.
- `internal/deploy` no longer depends on `internal/config`; template commit pinned in `boilerplate-source.yaml`.

### Removed or renamed (operator-visible)

Applies to anyone upgrading from the **first** published **v0.3.0** (`46fa3de`) or any build before **v0.4.0**:

- `rtifact aws|azure|gcp …` command groups removed — use root commands (`rtifact apply`, `rtifact validate`, …); cloud comes from YAML `cloud:`.
- `dataops` command group removed — use deploy `protect` / engine flags (C2+).
- `validate --check-azure` renamed to **`validate --check-quota`** (default from cloud hooks when the flag is omitted).
- **`vars/versions.yml`** moved to **`build/provider-versions.yml`** (`make bump-providers`).

### Architect pass (W0–W2)

- Golden fixtures for merge, tfvars, and backends; stop tracking packaged example artifacts.
- Normalize paths and YAML helpers via `internal/maputil`; unified `cloud.Cloud` surface (drop `cloudconfig`).
- Slim CLI (root lifecycle only), cloud `Hooks` / lifecycle on providers, provider pins in `build/provider-versions.yml`.

### Refactor (W3–W5)

- `PrepareLayerVars` on `cloud.Lifecycle`; Azure RG injection in the Azure provider; runner imports `cloud` + `registry` only.
- `internal/engine` step orchestration and confirm prompts; `TofuRunner`; phased layer steps; `context` from CLI.
- Profile-owned checklists and YAML `aliases`; `ResolveProfileID` from packaged profiles; packaging loops `profiles/*/profile.yaml`.

### Managed sync (W6)

- Pin **BoilerplatesRef v3.2.1**; `boilerplate-source.yaml` records template commit; verify fails on moved tags.

### Tests and release (W7–W8)

- `internal/archtest` import rules and cloud-string guards; engine and hook wiring tests (fake `Cloud` + `TofuRunner`).
- ADR 0010: engine steps and cloud boundaries; release workflow `deploy validate` smoke.

## [0.3.0] - 2026-11-22

Aligned with **boilerplates [v3.2.0](https://github.com/cipherops-io/boilerplates/releases/tag/v3.2.0)**.

### Deploy (Step B + C1)

- `rtifact deploy` command group: `validate`, `init-config write` (plan/apply/status/destroy stubbed for C2+).
- Packaged `schemas/deploy.schema.json` and `examples/deploy/` for every cloud binary.
- Helm SDK linked; `rtifact version` prints `deploy-engine`.
- Kubeconfig also written under `~/.rtifact/<prefix>/` for deploy discovery.
- `validate --explain` with `${env.*}` masking; placeholder checks in `values_files`; `crds` defaults to `chart` on load; `--kubeconfig` relative paths use CWD.
- CI: `deploy-unit` (not cluster e2e); `build-binary` prints binary size.
- Pre-C2: Helm **v4.3.2** SDK (Go 1.26); chart-style `${...}` allowed; shared env placeholder regex; defaults for wait/rollback; `ResolvedSecretRedactor` stub for C2.

### Architect pass (pre-C2)

- Remove `dataops` CLI stubs; root lifecycle commands only (no `rtifact aws|azure|gcp` duplicate groups).
- `internal/maputil` for shared map/YAML helpers; deploy tree walks unified in `walk.go`.
- Cloud `Hooks` (schema/maturity/quota) and `LifecycleHooks` (AWS destroy guard, Route53 delegation) on providers.
- Stray paths: `build/provider-versions.yml`, `build/scripts/github-e2e-prepare-config.sh`; drop tracked `local.mk`, `vars/versions.yml`, `docs/layer-contracts.md`.
- `validate --check-quota` (replaces `--check-azure`); default from cloud hooks when flag omitted.

## [0.2.0] - 2026-11-22

Aligned with **boilerplates [v3.1.1](https://github.com/cipherops-io/boilerplates/releases/tag/v3.1.1)**.

### Config and schema

- Single customer config **schema_version: 1** (version 2 removed).
- Explicit **features** block required in customer YAML; paste-ready hint built from `profiles/<profile>/<cloud>/features.yaml`.
- `init-config` copies `examples/<cloud>/minimal.yaml` with tier and customer filled in.

### Cloud layout (v3.1 work)

- Per-cloud profiles, schemas, contracts, default-off features, per-customer `build-binary` workflow.
- AWS validation (closed JSON Schema, preflight, destroy guard for EC2/ELB/EBS), DNS delegation verify (Route53 + fail-closed).

### AWS EKS

- Node group defaults in `profiles/aios-infra/aws/defaults.yaml`; module uses required fields with empty-list fallbacks only.

### Examples

- Tata Digital (`examples/ai-os-centralindia/config.yaml`) and Azure e2e configs carry explicit feature blocks for default-off profiles.