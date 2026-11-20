<!-- DO NOT EDIT — managed by cipherops-io/boilerplates (rtifact-product).
     Change it there, then run: make boilerplate-update -->
# AGENTS.md — modules (aws)

Modules wired in this cloud (from `boilerplate-vars.yml` → `Modules`):

- `compute/`
- `data/`
- `eks/`
- `naming/`
- `private-endpoint/`
- `security/`
- `state-backend/`
- `vpc/`


Add a module: scaffold with `make scaffold-module CLOUD=aws MODULE=<name>`, then register in `boilerplate-vars.yml` and profiles.