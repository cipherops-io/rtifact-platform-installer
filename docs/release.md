# Releasing rtifact

Every customer gets the same build of a version. There is no per-customer build. `.github/workflows/release.yml` (managed from boilerplates) builds one binary per cloud and platform, smoke-tests each binary on its own platform, then publishes them to Zot and signs them with cosign.

## Cutting a release

1. `main` is green, including `deploy-e2e`.
2. Optional dry run of `main`: `gh workflow run release.yml -f version=v0.1.0 -f ref=main -f publish=false`. This builds and smoke-tests everything and pushes nothing.
3. Add a `CHANGELOG.md` entry, then tag and push: `git tag v0.1.0 && git push origin v0.1.0`.
4. The tag push runs `release.yml`. The run summary and the GitHub release notes list every pushed digest.

To re-run a published version, dispatch with `version=v0.1.0` (no `ref`). This is a no-op when the content is unchanged.

| Input | Default | Meaning |
|-------|---------|---------|
| `version` | tag name | Tag to build. Must exist unless `ref` is set. |
| `ref` | empty | Build this ref instead. Only allowed with `publish=false`. |
| `clouds` | `azure,aws,gcp` | Clouds to build. |
| `publish` | `true` | `false` stops after the smoke tests. |
| `replace_existing` | `false` | Overwrite exact-version tags that hold different content. Only for the v0.0.1 period. |

## Jobs

- **verify**: checks the inputs, runs `make package-content go-test`, and validates the minimal example and deploy config for each cloud.
- **build**: runs `make release PACKAGED_CLOUDS=<cloud> GOOS=<os> GOARCH=<arch>` for each cloud × platform pair (linux/amd64, linux/arm64, darwin/arm64, darwin/amd64). Each binary embeds OpenTofu and every layer's providers for its platform, and `check-release.sh` rejects a binary that doesn't.
- **smoke**: runs on each platform's own runner, in an empty directory with no network registry access for providers: `version -o json`, `doctor --deep` (offline `tofu init` of every layer), `init-config`, `validate`, `deploy init-config` and `deploy validate`.
- **publish**: runs `make publish` (`build/scripts/publish-zot.sh`), then attaches `SHA256SUMS`, `cosign.pub` and `install.sh` to the GitHub release.

## Registry layout

Repository: `zot.prod.rtifact.ai/rtifact-installer`. Artifact type and layer media type: `application/vnd.rtifact.installer.v1`. The layer title is `rtifact`.

| Tag | Content | Mutable |
|-----|---------|---------|
| `v0.1.0-aws-linux-amd64` | One binary | No |
| `v0.1.0-aws` | Index of the four platform artifacts for the cloud | No |
| `v0.1.0-checksums` | `SHA256SUMS` for all binaries of the version | No |
| `latest-aws` | Same index as the newest non-prerelease `vX-aws` | Moves on each release |

Each artifact carries `org.opencontainers.image.{version,revision,source,created}`, `ai.rtifact.cloud`, `ai.rtifact.platform` and `ai.rtifact.tofu-version` annotations.

Builds are reproducible (`-trimpath`, version stamped from the tag). Before pushing anything, `publish-zot.sh` compares each existing tag's binary digest with the new build:

- Same content: skipped, so a re-run changes nothing.
- Different content: the whole run fails before any push, unless `replace_existing=true`.
- Prerelease versions (`v1.0.0-rc.1`) never move `latest-<cloud>`.

## Signing

Every artifact, index and checksums manifest is signed with the cosign key in the `COSIGN_PRIVATE_KEY` / `COSIGN_PASSWORD` secrets, with no public transparency log upload. The matching public key is `build/cosign.pub`, and publishing fails if the secret key doesn't match it. The private key backup lives outside the repo (ops `Zot/cosign/`).

## Getting a binary

With the install script (oras if installed, otherwise curl; checks the binary against the registry digest):

```bash
export RTIFACT_REGISTRY_USERNAME=<user> RTIFACT_REGISTRY_PASSWORD=<password>
sh install.sh --cloud aws                        # latest-aws for this machine
sh install.sh --cloud aws --version v0.1.0 --key cosign.pub --dir /usr/local/bin
```

By hand:

```bash
oras login zot.prod.rtifact.ai
oras pull --platform linux/amd64 zot.prod.rtifact.ai/rtifact-installer:latest-aws
oras pull zot.prod.rtifact.ai/rtifact-installer:v0.1.0-aws-darwin-arm64
cosign verify --key cosign.pub --insecure-ignore-tlog=true zot.prod.rtifact.ai/rtifact-installer:v0.1.0-aws-darwin-arm64
chmod +x rtifact && ./rtifact doctor
```

## Settings

| Where | Name | Value |
|-------|------|-------|
| Repo variable | `ZOT_HOST` | `zot.prod.rtifact.ai` |
| Repo variable | `ZOT_USERNAME` | `github-ci` (push access to `rtifact-installer`) |
| Secret | `ZOT_PASSWORD` | Password for `ZOT_USERNAME` |
| Secret | `COSIGN_PRIVATE_KEY` | Contents of `cosign.key` |
| Secret | `COSIGN_PASSWORD` | Password for `cosign.key` |

Customers need a Zot user with read access to `rtifact-installer`.

## Local

```bash
make release PACKAGED_CLOUDS=aws GOOS=linux GOARCH=arm64 VERSION=v0.1.0   # dist/rtifact
./dist/rtifact doctor --deep
RTIFACT_TEST_ZOT_BIN=/path/to/zot go test ./build/scripts/tests -run TestPublishZot   # publish + install.sh against a local Zot
```
