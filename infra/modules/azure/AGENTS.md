<!-- DO NOT EDIT — managed by cipherops-io/boilerplates (rtifact-product).
     Change it there, then run: make boilerplate-update -->
# AGENTS.md — modules (azure)

Modules wired in this cloud (from `boilerplate-vars.yml` → `Modules`):

- `ai-foundry/`
- `aks/`
- `app-gateway/`
- `bastion-host/`
- `compute/`
- `connectivity/`
- `data/`
- `dns-public/`
- `firewall/`
- `naming/`
- `network/`
- `private-endpoint/`
- `security/`
- `state-backend/`


Add a module: scaffold with `make scaffold-module CLOUD=azure MODULE=<name>`, then register in `boilerplate-vars.yml` and profiles.