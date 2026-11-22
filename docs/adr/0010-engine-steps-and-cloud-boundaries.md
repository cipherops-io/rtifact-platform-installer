# ADR 0010: Engine steps and cloud boundaries

## Status

Accepted

## Context

v0.4 refactors moved layer apply orchestration into `internal/engine`, cloud-specific
layer preparation into `cloud.Lifecycle`, and thinned the runner so it depends on
the neutral `cloud` surface rather than per-provider packages.

## Decision

1. **Orchestration:** The installer executes ordered work as `engine.Step` values
   (`Run`, `ConfirmIfNeeded`). Runner layer work is phased: guards, workspace
   (vars, backend, tofu runner), then tofu CLI actions.
2. **Cloud hooks:** `PrepareLayerVars`, DNS delegation, destroy orphans, and
   preflight/schema defaults live on `cloud.Cloud` implementations. Callers outside
   `internal/cloud` resolve providers through `cloud/registry` only.
3. **Import rules:** `internal/runner` imports `internal/cloud` and
   `internal/cloud/registry` only (not `cloud/aws`, `cloud/azure`, or `cloud/gcp`).
   Other packages must not import provider subpackages; only `registry` wires them.
4. **String compares:** Literal comparisons to `azure`, `aws`, or `gcp` outside
   `internal/cloud` are discouraged; configuration merge and CLI defaults remain
   the narrow exceptions documented in `internal/archtest`.
5. **Regression guards:** `internal/archtest` and focused runner/engine tests lock
   these boundaries in CI (`go test ./...`).

## Consequences

- New cloud behaviour is added on provider types and registered in `registry`.
- Runner changes stay orchestration-only; guards and confirms stay in engine.
- Golden fixtures (W0) remain the behaviour oracle across refactors.