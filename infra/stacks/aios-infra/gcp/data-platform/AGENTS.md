<!-- DO NOT EDIT — managed by cipherops-io/boilerplates (rtifact-product).
     Change it there, then run: make boilerplate-update -->

# AGENTS.md — data-platform (gcp)

## Purpose

Stateful data, secrets, registries, private endpoints and private DNS.

**Put here if:** It holds data, secrets, keys, images or files, or exposes a private endpoint or private DNS for such a service.

**Not here:**

- Kubernetes → app-platform
- Public gateway or WAF → network-edge
- Core VNet routing → core-network

**Examples:**

- Database or cache
- Object storage or file share
- Key vault or KMS
- Container registry
- Private endpoint and private DNS for PaaS

## Modules in use today

- `data`

- `security`

- `private-endpoint`


**Reads from:** core-network

**Provides:** `docs/layer-contracts.md#data-platform`

**How to place a new service:** `infra/AGENTS.md#placement`