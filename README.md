# rtifact-platform-installer

Go CLI that embeds OpenTofu and deploys layered Azure platform stacks from customer YAML.

## Commands

```bash
make build-dev
INSTALLER_CONTENT_ROOT=. ./dist/installer validate examples/sandbox.yaml --check-azure=false
INSTALLER_CONTENT_ROOT=. ./dist/installer plan examples/sandbox.yaml --stack network --yes
INSTALLER_CONTENT_ROOT=. ./dist/installer apply examples/sandbox.yaml --stack all
INSTALLER_CONTENT_ROOT=. ./dist/installer bootstrap examples/sandbox.yaml --yes
INSTALLER_CONTENT_ROOT=. ./dist/installer preflight examples/sandbox.yaml
./dist/installer data backup config.yaml
```

Layers: `bootstrap`, `network`, `data`, `cluster` (each with its own state file when using remote backend).

Module: `github.com/cipherops-io/rtifact-platform-installer`

Generated from [`cipherops-io/boilerplates`](../boilerplates) `devops-templates/platform-installer`.