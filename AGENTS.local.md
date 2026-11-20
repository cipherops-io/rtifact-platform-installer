# AGENTS.local.md — installer-specific (not managed)

## Feature flags

- New IaC must be gated by `features.<name>` in the active profile.
- Update `profiles/<id>/features.yaml`, `schemas/`, and the profile registry when adding flags.

## OIDC / E2E

- Live Azure deploy: `.github/workflows/azure-e2e.yml` (workflow_dispatch).
- Examples: `examples/github-e2e-azure.yaml`

## Provider bumps

```bash
make bump-providers   # re-renders infra/stacks/**/versions.tofu from build/boilerplate/stack-versions/layer-versions
make tofu-lock
```

Provider and extra-provider pins: `vars/versions.yml`. OpenTofu semver and registry `ProviderSources` come from managed `tooling.mk` (boilerplates `defaults/tool-versions.yml`).