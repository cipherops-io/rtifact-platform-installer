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

Var files: `boilerplate-vars.yml`, `../boilerplates/vars/versions.yml`, `../boilerplates/vars/layers.yml`.

## Before you push

Switch away from a local template path — CI runs `make check-template-source`:

```make
BOILERPLATES_TEMPLATE_URL = github.com/cipherops-io/boilerplates//devops-templates/rtifact-product?ref=v3.0.0
```

## Verify

```bash
make boilerplate-verify
```