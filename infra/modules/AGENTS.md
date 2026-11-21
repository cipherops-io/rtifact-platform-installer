<!-- DO NOT EDIT — managed by cipherops-io/boilerplates (rtifact-product).
     Change it there, then run: make boilerplate-update -->
# AGENTS.md — modules

Per-module conventions: `providers.tofu`, `variables.tofu`, `main*.tofu`, `outputs.tofu`, README with `BEGIN_TF_DOCS` markers.

See `infra/modules/<cloud>/AGENTS.md` for per-cloud module inventory.

## Cloud layout rules

1. Modules live under `infra/modules/<cloud>/` only; do not mix cloud providers or cross-reference another cloud's module path.
2. Expose optional behaviour via stack-level `features.<name>` only (default false); modules should not read `enabled` fields from nested config objects.
3. Module variables are wiring and shape — not customer defaults. Customer defaults stay in `profiles/<id>/<cloud>/defaults.yaml`.