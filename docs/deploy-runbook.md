# Deploy runbook

Operator guide for `rtifact deploy` on a customer cluster. Always run `rtifact deploy validate <file>` (and `validate --explain` when reviewing secrets) before `plan` or `apply`.

## Prerequisites

- Kubeconfig path in the deploy file or `--kubeconfig` (relative paths resolve from the deploy file directory).
- OCI registry credentials via `registry.auth.*_env` when charts are private.
- Deployment ID (`customer.code` + `environment`) stable across applies — it names inventory, locks, and Helm labels.

## Command reference

| Command | Purpose | Notes |
|---------|---------|--------|
| `deploy validate` | Schema + semantic checks | Refuses cycles, unknown `requires`, missing `values_files`, bad chart refs |
| `deploy validate --explain` | Human summary with `${env.*}` masked | No cluster contact |
| `deploy plan` | Read-only diff | Exit **2** when changes pending; refusals exit **1** |
| `deploy plan -o json` | Machine-readable plan | Use in CI with `--detailed-exitcode` |
| `deploy apply` | Mutating sync | Prompts unless `--yes` |
| `deploy status` | Inventory vs live Helm | `status -o json` for automation |
| `deploy rollback <release>` | Roll back one stuck/pending release | After plan shows recovery hint |
| `deploy destroy` | Uninstall all inventory releases | PVCs kept unless `--purge` |

## Typical workflow

1. **Validate** locally: `rtifact deploy validate ./my-env.yaml`
2. **Plan** against the target cluster: `rtifact deploy plan ./my-env.yaml --detailed-exitcode`
3. **Apply** after review: `rtifact deploy apply ./my-env.yaml` (add `--yes` only in locked automation)
4. **Status** after incidents: `rtifact deploy status ./my-env.yaml`

## Identity and safety

- Set `cluster.expected_server` to the API server URL you intend (identity guard runs before writes).
- Protected releases (`protect: true`) are not removed unless `deploy destroy --allow-destroy` (or apply with the matching flag).
- Downgrades are refused unless `--allow-downgrade`.
- Only one apply per deployment ID at a time; a second apply receives “Lock held by …”. Use `--force-unlock` only when the holder process is dead.

## Registry and pull secrets

- `registry.plain_http` / `ca_file` for insecure or private CAs (common in lab Zot setups).
- Namespace `pull_secret: true` creates/refreshes an image pull secret and optionally attaches it to the default service account.

## CI

- Unit tests: `make deploy-unit` (no cluster).
- Cluster proof: `make deploy-e2e` (kind + Zot). Scenario coverage is listed in `tests/deploy/README.md` (S1–S19).

## Troubleshooting

| Symptom | Check |
|---------|--------|
| Plan exit 1, “refused” | Mass removal, downgrade, protect, or stuck dependent — read plan notes |
| Apply failed mid-run | `deploy status`; inventory lists what committed; `rollback` for pending release |
| Wrong cluster | `expected_server` mismatch — fix kubeconfig or server URL |
| Chart pull failed | Registry auth env vars, host, `plain_http`, network |
| Empty inventory after refused apply | Expected — refusals do not write inventory |

## Examples

- `examples/deploy/demo/demo.yaml` — same release layout as the e2e fixture (`ok-a` / `ok-b` with `requires`), with placeholder registry host.
- `examples/deploy/minimal.yaml` — smallest valid file.
- `examples/deploy/reference/` — multi-release sample with external values file.