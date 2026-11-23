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
| `deploy` (root) | Orchestration: session, plan, apply, destroy, rollback, status; no client-go |
| `plan` | Decisions only; no Helm/Kubernetes imports |
| `model` | Deploy file types; standard library only |
| `cluster` | Namespaces, PVCs, pull secrets on the session clientset |
| `helmclient` | Only importer of `helm.sh/helm/v4`; `helmclient/helmfake` for unit tests |
| `kube`, `ocireg`, `crds`, `inventory`, `lock`, `redact` | As named |

`internal/archtest` enforces: `plan` has no `helm.sh`/`k8s.io`; only `helmclient` imports Helm; only `kube`, `cluster`, `lock`, `inventory`, `crds` and `helmclient` import `k8s.io`; `model` imports only the standard library; deploy must not import `internal/runner`, `internal/config`, or `internal/cloud`.

Tests: unit tests use `k8s.io/client-go/kubernetes/fake` and `helmfake`; cluster scenarios are `make deploy-e2e` (see `tests/deploy/README.md`).
