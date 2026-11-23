# Changelog

## [0.0.1] - 2026-11-23

First public release, paired with **boilerplates [v0.0.1](https://github.com/cipherops-io/boilerplates/releases/tag/v0.0.1)** (`c4c616a` in `boilerplate-source.yaml`). Earlier internal version numbers (0.2–0.4.x) were pre-release builds and are folded into this entry. An earlier `v0.0.1` tag pointed at a build whose deploy commands did not reach the cluster; this tag replaces it.

### Deploy engine

`rtifact deploy validate | plan | apply | status | rollback | destroy` with Helm v4.3.2 (SDK), an inventory ConfigMap per deployment, a Lease lock held from plan through apply, and stepped apply in dependency order.

`deploy-e2e` (kind cluster plus authenticated Zot registry) was first green on `main` at `41f1759` ([CI run 45121102246](https://github.com/cipherops-io/rtifact-platform-installer/actions/runs/45121102246)); every CI check, `deploy-e2e` included, is green at `8d01e55` ([CI run 45129873664](https://github.com/cipherops-io/rtifact-platform-installer/actions/runs/45129873664)). It runs the Go scenarios in `tests/deploy/README.md` and `tests/deploy/rehearsal.sh` against the built binary:

- **S1–S4**: fresh apply, no-op re-apply, version bump, value-only change
- **S5–S6**: failed release rollback, CRD upgrade (`crds: upgrade`)
- **S7–S10**: removal, required-dependency guard, protect refusal, mass-removal guard
- **S11–S12**: stuck (pending) release recovery via `rollback`, concurrent apply lock
- **S13–S17**: identity guard, invalid values file, pull-secret refresh on credential rotation, downgrade guard, bad registry password
- **S18**: destroy keeps PVCs
- **Extras**: secret redaction canary, adopting an unmanaged release, interrupted apply (exit 130, lock released)

Plan exit codes: `0` no changes, `1` refusal or error (a refusal always wins), `2` pending changes with `--detailed-exitcode`, `130` interrupted. `plan -o json` and `status -o json` carry `schema_version`.

### Infrastructure

- One `cloud.Cloud` interface for Azure, AWS and GCP; shared `internal/engine` plan/confirm/act/record path; profiles are data under `profiles/<id>/`.
- Customer config `schema_version: 1` with an explicit `features` block; `init-config` copies `examples/<cloud>/minimal.yaml`.
- AWS validation (closed JSON Schema, preflight, destroy guard), Route53 delegation verify, EKS `pod_identities` (preview).
- AWS `data-platform` layer (preview features): `kms`, `postgresql` (RDS, Secrets Manager-managed master password), `redis` (ElastiCache, TLS and at-rest encryption), `object_storage` (S3) and `container_registry` (ECR). An enabled `postgresql` or `redis` must name a `subnet_tier` spanning two zones. Checked with `tofu validate` and mocked-provider tests; not yet applied to a live account.
- Provider pins in `build/provider-versions.yml`; managed files and `boilerplate-source.yaml` from boilerplates.

### Upgrading from pre-release builds

- `rtifact aws|azure|gcp …` command groups are removed: use root commands (`rtifact apply`, `rtifact validate`, …); the cloud comes from YAML `cloud:`.
- `dataops` command group is removed: use deploy `protect` and the engine flags.
- `validate --check-azure` is now `validate --check-quota`.
- `vars/versions.yml` moved to `build/provider-versions.yml` (`make bump-providers`).
