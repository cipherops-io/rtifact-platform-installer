# Adding a profile

1. **Same Azure topology, different defaults**  
   - Add `profiles/<id>/defaults.yaml` and `profile.yaml` with existing `stack_id`.

2. **New topology**  
   - Add modules under `infra/modules/` if needed.
   - Add `infra/stacks/<id>/<cloud>/<layer>/` compositions for each layer (`providers.tofu` + `*.tofu` only; see `AGENTS.md`).
   - Point new `profiles/<id>/profile.yaml` `stack_id` at the stack and list supported `clouds`.

3. **Customer contract**  
   - Extend `schemas/customer-config.schema.json` only if new top-level keys are required.
   - Set default `features` and sizing in `profiles/<id>/defaults.yaml`; customers override via top-level sections (replace merge for `features` and `kubernetes_cluster.node_pools`).

4. **Release**  
   - Run `make package-content` before `go build`.
   - Refresh provider mirror: `make mirror-tofu`.

5. **Checklist**  
   - Add `internal/checklist/<id>.md` and register in `internal/checklist/checklist.go`.