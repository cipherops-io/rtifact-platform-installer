<!-- DO NOT EDIT — managed by cipherops-io/boilerplates (rtifact-product).
     Change it there, then run: make boilerplate-update -->
# AGENTS.md — deploy

1. To add an operator: one entry in the releases list. Never Go code.
2. Pin an exact version. Never a range.
3. List what it `requires`. A custom resource always requires its operator.
4. Set `crds: upgrade` for any chart that ships CRDs.
5. Mark anything that holds data `protect: true`.
6. No secret values in the file; use `${env.NAME}`.
7. A raw manifest becomes a small chart; it is never applied directly.
8. A new field needs the schema, the example and a test in the same change.

## Package layout (`internal/deploy/`)

| Package | Role |
|---------|------|
| `plan` | Decisions only; no Helm/Kubernetes imports |
| `apply` | Lock, namespaces, CRDs, releases, inventory |
| `helmclient` | Only importer of `helm.sh/helm/v4` |
| `kube`, `ocireg`, `crds`, `inventory`, `lock`, `redact` | As named |

`internal/archtest` enforces: `plan` has no `helm.sh`/`k8s.io`; only `helmclient` imports Helm; deploy must not import `internal/runner`, `internal/config`, or `internal/cloud`.
