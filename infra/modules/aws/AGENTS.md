<!-- DO NOT EDIT — managed by cipherops-io/boilerplates (rtifact-product).
     Change it there, then run: make boilerplate-update -->
# AGENTS.md — modules (aws)

Registered modules (from `boilerplate-vars.yml` → `Modules`), grouped by layer:

### state-backend

- `state-backend/`


### core-network

- `naming/`
- `vpc/`
- `bastion/`


### data-platform

- `data/`
- `security/`
- `private-endpoint/`


### app-platform

- `eks/`
- `compute/`
- `karpenter-infra/`
- `pod-identity/`


### network-edge

- `dns-public/`
- `certificates/`


Add a module: `make scaffold-module CLOUD=aws MODULE=<name>`, register as `<layer>/<name>` in `boilerplate-vars.yml`, add `features.<name>` to the profile.