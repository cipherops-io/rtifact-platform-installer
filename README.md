# rtifact-platform-installer

Go CLI **`rtifact`** embeds OpenTofu and deploys layered **aios-infra** stacks on Azure, AWS, and GCP from customer YAML.

## Commands

```bash
make build-dev
INSTALLER_CONTENT_ROOT=. ./dist/rtifact validate examples/sandbox.yaml --check-azure=false
INSTALLER_CONTENT_ROOT=. ./dist/rtifact azure plan examples/sandbox.yaml --stack core-network --yes
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

All artifacts are named `dist/rtifact` with only the packaged cloud subcommands registered.

## Azure E2E (GitHub Actions + OIDC)

Live **validate + plan** in Azure uses [`.github/workflows/azure-e2e.yml`](.github/workflows/azure-e2e.yml) (manual dispatch; defaults: all stacks, quota check on). Setup: [docs/azure-github-oidc-e2e.md](docs/azure-github-oidc-e2e.md). Local apply/destroy: `make e2e-azure-deploy` with `E2E_AZURE_CONFIG`, `E2E_AZURE_ACTION`, `E2E_AZURE_STACK`.

Module: `github.com/cipherops-io/rtifact-platform-installer`

Generated from [`cipherops-io/boilerplates`](../boilerplates) `devops-templates/platform-installer`.