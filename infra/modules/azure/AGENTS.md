<!-- DO NOT EDIT — managed by cipherops-io/boilerplates (rtifact-product).
     Change it there, then run: make boilerplate-update -->
# AGENTS.md — modules (azure)

Registered modules (from `boilerplate-vars.yml` → `Modules`), grouped by layer:

### state-backend

- `state-backend/`


### core-network

- `naming/`
- `network/`
- `firewall/`
- `bastion-host/`
- `connectivity/`


### data-platform

- `data/`
- `security/`
- `private-endpoint/`


### app-platform

- `aks/`
- `compute/`


### network-edge

- `app-gateway/`
- `dns-public/`


### ai-integrations

- `ai-foundry/`


Add a module: `make scaffold-module CLOUD=azure MODULE=<name>`, register as `<layer>/<name>` in `boilerplate-vars.yml`, add `features.<name>` to the profile.