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

## Per-customer binary build (`build-binary.yml`)

Workflow dispatch on a **release tag** (not branch HEAD):

```bash
make release PACKAGED_CLOUDS=<cloud> VERSION=<tag> CUSTOMER=<name> GOOS=<os> GOARCH=<arch>
```

Produces `dist/<ProductName>`. Smoke checks:

```bash
./dist/<ProductName> version   # prints version, packaged cloud(s), customer
./dist/<ProductName> validate examples/<cloud>/minimal.yaml
./dist/<ProductName> deploy validate examples/deploy/minimal.yaml
```

Deploy content is packaged in every binary, regardless of the selected infrastructure cloud.

Managed workflow: `.github/workflows/build-binary.yml` (matrix targets from template `BuildTargets`).

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