<!-- DO NOT EDIT — managed by cipherops-io/boilerplates (rtifact-product).
     Change it there, then run: make boilerplate-update -->

# Boilerplate sync (rtifact-product)

## Update managed files

```bash
make boilerplate-update
```

Local template (sibling checkout):

```bash
make boilerplate-update BOILERPLATES_TEMPLATE_URL=../boilerplates/devops-templates/rtifact-product
```

Var files: `boilerplate-vars.yml` only. Tool pins and `LayerMeta` come from `boilerplate.yml` variable defaults in the tagged product template (no extra var-file copies).

## Before you push

CI uses the GitHub template ref (not a local path). Override only for local work:

```make
BOILERPLATES_TEMPLATE_URL = github.com/cipherops-io/boilerplates//devops-templates/rtifact-product?ref=v0.0.1
```

## Release build and publish (`release.yml`)

One build per version for all customers: one binary per cloud and platform (`Clouds` × `BuildTargets`), pushed to `zot.prod.rtifact.ai/rtifact-installer`. The workflow runs on a `v*` tag push and on dispatch, and calls only `make` targets and the built binary:

```bash
make release PACKAGED_CLOUDS=<cloud> VERSION=<tag> GOOS=<os> GOARCH=<arch>   # dist/rtifact, contents checked
make publish PUBLISH_DIR=out VERSION=<tag> REGISTRY=<host> REGISTRY_REPOSITORY=<repo> PACKAGED_CLOUDS=<clouds>
```

Smoke checks run on each platform's own runner, in an empty folder:

```bash
./rtifact version -o json    # version, commit, clouds, mode=bundled, platform, tofu_version
./rtifact doctor --deep      # offline tofu init of every layer from the embedded mirror
```

Deploy content is packaged in every binary, regardless of the selected infrastructure cloud. Repository settings: variables `ZOT_HOST`, `ZOT_USERNAME`; secrets `ZOT_PASSWORD`, `COSIGN_PRIVATE_KEY`, `COSIGN_PASSWORD`.

## Removed managed files

Boilerplate does not delete a file that left the template. `make boilerplate-verify` runs `check-managed-orphans`, which fails on tracked files that carry the managed header but are not in `boilerplate-manifest.yaml`. Delete them in the same change as the update.

| Removed | Replaced by |
|---|---|
| `.github/workflows/build-binary.yml` | `.github/workflows/release.yml` |

## Template source pin (`boilerplate-source.yaml`)

After every `make boilerplate-update`, `record-boilerplate-source` writes `ref`, `commit`, and `template_url` to **`boilerplate-source.yaml`** (created by that step, not re-rendered from the template). Commit it with the sync so CI can detect a **moved tag** (same ref name, different commit).

`verify-boilerplate-source` runs inside `make boilerplate-verify` for GitHub template URLs. It is **skipped** when `BOILERPLATES_TEMPLATE_URL` points at a local sibling checkout (`../boilerplates/...`).

## Verify

CI checks managed files against `boilerplate-manifest.yaml` checksums (no cross-repo fetch):

```bash
go test ./tests/config/... -run TestManagedManifestChecksums
```

Optional local re-render:

```bash
make boilerplate-verify
make deploy-unit
```

Commit or stash local changes before running verify.