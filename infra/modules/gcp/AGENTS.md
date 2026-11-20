<!-- DO NOT EDIT — managed by cipherops-io/boilerplates (rtifact-product).
     Change it there, then run: make boilerplate-update -->
# AGENTS.md — modules (gcp)

Registered modules (from `boilerplate-vars.yml` → `Modules`), grouped by layer:

### state-backend

- `state-backend/`


### core-network

- `naming/`
- `vpc/`


### data-platform

- `data/`
- `security/`
- `private-endpoint/`


### app-platform

- `gke/`
- `compute/`


Add a module: `make scaffold-module CLOUD=gcp MODULE=<name>`, register as `<layer>/<name>` in `boilerplate-vars.yml`, add `features.<name>` to the profile.