# ADR 0011: Deploy engine architecture

## Status

Accepted (v0.0.1)

## Context

`rtifact deploy` must plan and apply Helm releases from versioned OCI charts on customer clusters, with the same identity and confirmation patterns as the infrastructure runner. The first **v0.0.1** tag shipped layout and CLI stubs before cluster behaviour was proven; this ADR records the design that passed `make deploy-e2e`.

## Decision

### Package boundaries

- **Pure planning** lives in `internal/deploy/plan` — no `helm.sh` or `k8s.io` imports; six removal/install guards are table-tested in `plan_test.go`.
- **Helm SDK** is confined to `internal/deploy/helmclient` (per-namespace `action.Configuration`, single label `rtifact.ai/deployment`, PVC keep via post-render).
- **Cluster access** uses explicit kubeconfig paths via `internal/deploy/kube`; identity (`expected_server`) is checked before reads or writes.
- **Registry auth** is in-memory from env vars (`internal/deploy/ocireg`); resolved literals are redacted via `internal/deploy/redact`.
- **Orchestration** (`internal/deploy` root) opens a **`Session`** (validate → load → discover → identity), then runs plan or stepped apply.

### Session (F1)

`OpenSession` is the single constructor for deploy commands. It returns `Document`, discovery paths, `kube.Config`, and `IdentityInfo`. OCI registry clients are created only when `RunOptions` needs chart pulls (plan/apply).

### Inventory v2 (F3/F4)

Inventory is a versioned JSON document in ConfigMap `rtifact-deploy-inventory` (namespace `rtifact-system`), keyed by deployment ID (`customer.code` + `environment`). Each entry records release name, namespace, chart ref, semver, chart digest, values digest, `requires`, and timestamps. Unknown newer inventory versions are refused on read.

### Apply steps (F4/F5)

Apply uses `engine.Step` ordering:

1. Ensure `rtifact-system` and release namespaces (plus pull secrets when configured).
2. Acquire coordination lease (deploy lock).
3. For each release in install order: optional CRD prelude → Helm install/upgrade → incremental inventory write.
4. Process **removals from plan rows** (not from the raw file list).
5. Release lock (including on interrupt where safe).

Destroy walks inventory removal order; rollback is per-release.

### Plan guards

`plan.Build` applies six guards (orphan removal, dependents, protect, mass removal, data retention, stuck/pending) documented in `tests/deploy/README.md` scenarios S7–S11 and S16.

### Helm client

One Helm `Client` holds per-namespace configurations so releases in different namespaces do not share a stale namespace context. Chart location uses an explicit temp cache root; ambient `HELM_*` env vars are not used for cluster targeting.

## Consequences

- Import rules are enforceable via `internal/archtest`.
- E2E scenarios S1–S19 (plus canary/adopt/interrupt extras) run against kind + Zot in `make deploy-e2e`.
- Infra placeholders `${infra.*}` resolve from discovered stack outputs when `outputs_dir` is configured.
- Examples under `examples/deploy/` mirror the e2e fixture shape for docs; production chart names remain customer-specific.