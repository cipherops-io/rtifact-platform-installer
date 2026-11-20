# AGENTS.local.md — installer-specific (not managed)

## Feature flags

- New IaC must be gated by `features.<name>` in the active profile.
- Update `profiles/<id>/features.yaml`, `schemas/`, and the profile registry when adding flags.

## OIDC / E2E

- Live Azure deploy: `.github/workflows/azure-e2e.yml` (workflow_dispatch).
- Examples: `examples/github-e2e-azure.yaml`

## Provider bumps

```bash
make bump-providers   # edits stack versions.tofu
make tofu-lock
```

Pins canonical file: `vars/versions.yml` in this repo.