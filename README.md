# rtifact-platform-installer

Go CLI **`rtifact`** embeds OpenTofu and deploys layered **aios-infra** stacks on Azure, AWS, and GCP from customer YAML.

## Commands

```bash
make build-dev
INSTALLER_CONTENT_ROOT=. ./dist/rtifact validate examples/azure/sandbox.yaml --check-quota=false
INSTALLER_CONTENT_ROOT=. ./dist/rtifact plan examples/azure/sandbox.yaml --stack core-network --yes
INSTALLER_CONTENT_ROOT=. ./dist/rtifact bootstrap examples/azure/sandbox.yaml --yes
```

Layers: `state-backend` (via `bootstrap`), then `core-network`, `data-platform`, `app-platform`, `network-edge`, `ai-integrations` (`--stack all`).

## Release binaries

Pushing a `v*` tag publishes one binary per cloud and platform to `zot.prod.rtifact.ai/rtifact-installer`, signed with cosign. See [docs/release.md](docs/release.md) for the tags and for `scripts/install.sh`.

Local builds embed OpenTofu and every provider for one platform:

| Target | `PACKAGED_CLOUDS` |
|--------|-------------------|
| `make release-all` | azure,aws,gcp |
| `make release-azure` | azure |
| `make release-aws` | aws |
| `make release-gcp` | gcp |

Add `GOOS=linux GOARCH=arm64` to cross-build. The output is always `dist/rtifact`; check it with `./dist/rtifact doctor --deep`.

Module: `github.com/cipherops-io/rtifact-platform-installer`

Managed layout from [`cipherops-io/boilerplates`](../boilerplates) `devops-templates/rtifact-product` (`make boilerplate-update`).

**CI:** GitHub Actions should require jobs `config-tests`, `tofu-fast`, and `go` on `main`. Managed-file drift is enforced by `go test ./tests/config/... -run TestManagedManifestChecksums` (no cross-repo token).