# rtifact-platform-installer

Go CLI **`rtifact`** embeds OpenTofu and deploys layered **aios-infra** stacks on Azure, AWS, and GCP from customer YAML.

## Commands

```bash
make build-dev
INSTALLER_CONTENT_ROOT=. ./dist/rtifact validate examples/azure/sandbox.yaml --check-quota=false
INSTALLER_CONTENT_ROOT=. ./dist/rtifact plan examples/azure/sandbox.yaml --stack core-network --yes
INSTALLER_CONTENT_ROOT=. ./dist/rtifact bootstrap examples/sandbox.yaml --yes
```

Layers: `state-backend` (via `bootstrap`), then `core-network`, `data-platform`, `app-platform`, `network-edge`, `ai-integrations` (`--stack all`).

## Release binaries

| Target | `PACKAGED_CLOUDS` |
|--------|-------------------|
| `make release-all` | azure,aws,gcp |
| `make release-azure` | azure |
| `make release-aws` | aws |
| `make release-gcp` | gcp |

All artifacts are named `dist/rtifact`; `cloud:` in customer YAML selects the target (build with `PACKAGED_CLOUDS` for smaller binaries).

Module: `github.com/cipherops-io/rtifact-platform-installer`

Managed layout from [`cipherops-io/boilerplates`](../boilerplates) `devops-templates/rtifact-product` (`make boilerplate-update`).

**CI:** GitHub Actions should require jobs `config-tests`, `tofu-fast`, and `go` on `main`. Managed-file drift is enforced by `go test ./tests/config/... -run TestManagedManifestChecksums` (no cross-repo token).