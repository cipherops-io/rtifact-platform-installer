# Deploy runbook

Use `rtifact deploy validate` before every plan. Run `plan` with `--detailed-exitcode` in CI. Machine-readable output: `deploy plan -o json` and `deploy status -o json`.

| Command | When |
|---------|------|
| `deploy plan` | Read-only diff; exit 2 when changes pending (refusals exit 1) |
| `deploy apply --yes` | Automation only with locked credentials |
| `deploy rollback` | Recover pending/stuck Helm revisions |
| `deploy destroy` | Tear down all releases in the file (PVCs kept) |

E2E: `make deploy-e2e` (kind + local OCI registry).