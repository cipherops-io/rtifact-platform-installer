# ADR 0011: Deploy engine packages and cluster I/O boundaries

## Status

Accepted (v0.5.0)

## Context

`rtifact deploy` must plan and apply Helm releases from versioned OCI charts while sharing confirmation and identity patterns with the infra runner.

## Decision

- **Pure planning** lives in `internal/deploy/plan` (no `helm.sh` or `k8s.io` imports).
- **Helm SDK** is confined to `internal/deploy/helmclient` (PVC keep via post-render).
- **Cluster access** uses explicit kubeconfig paths via `internal/deploy/kube`; identity is checked before reads.
- **Registry auth** is in-memory from env vars (`internal/deploy/ocireg`); writers use `internal/deploy/redact`.
- **Orchestration** (`internal/deploy` root) wires plan → confirm (`internal/engine` deploy helpers) → apply.

## Consequences

- Import rules are enforceable via `internal/archtest`.
- E2E scenarios (S1–S18) run against kind + Zot in `make deploy-e2e`.
- Infra placeholders `${infra.*}` resolve from discovered stack outputs when configured.