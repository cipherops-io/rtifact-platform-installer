# Operations backlog (post hardening)

Ordered follow-ups from the infrastructure hardening plan (section 8):

1. **Upgrade path** — record installer version per layer; refuse downgrade; `rtifact upgrade` plans all stacks.
2. **Drift detection** — `rtifact status --drift` with plan exit codes per layer.
3. **Run lock** — single lock per customer prefix for concurrent `--stack all` runs.
4. **Air-gap bundle** — signed archive + `rtifact bundle verify`.
5. **Infra → app contract** — version outputs for Helm/operator consumers.
6. **Application install** — Helm driven by installer (depends on item 5).
7. **Support bundle** — `rtifact support-bundle` with redacted config and tool versions.

Phase 4 **A1** (lift hardcoded module literals to YAML) remains incremental; see `TestModulesSensitiveLiteralsDocumentedInVars` log output.