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

Var files: `boilerplate-vars.yml` only, plus template-bundled defaults (`defaults/layers.yml`, `defaults/tool-versions.yml`) resolved automatically from the same git ref as the product template.

## Before you push

CI uses the GitHub template ref (not a local path). Override only for local work:

```make
BOILERPLATES_TEMPLATE_URL = github.com/cipherops-io/boilerplates//devops-templates/rtifact-product?ref=v3.0.0
```

## Verify

Re-renders managed files and fails if anything drifted (manifest excluded):

```bash
make boilerplate-verify
```

Commit or stash local changes before running verify.