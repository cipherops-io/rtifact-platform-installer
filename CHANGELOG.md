# Changelog

## [0.3.0] - 2026-11-22

Aligned with **boilerplates [v3.2.0](https://github.com/cipherops-io/boilerplates/releases/tag/v3.2.0)**.

### Deploy (Step B + C1)

- `rtifact deploy` command group: `validate`, `init-config write` (plan/apply/status/destroy stubbed for C2+).
- Packaged `schemas/deploy.schema.json` and `examples/deploy/` for every cloud binary.
- Helm SDK linked; `rtifact version` prints `deploy-engine`.
- Kubeconfig also written under `~/.rtifact/<prefix>/` for deploy discovery.

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