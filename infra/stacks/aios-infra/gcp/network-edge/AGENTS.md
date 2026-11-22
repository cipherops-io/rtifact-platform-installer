<!-- DO NOT EDIT — managed by cipherops-io/boilerplates (rtifact-product).
     Change it there, then run: make boilerplate-update -->

# AGENTS.md — network-edge (gcp)

## Purpose

Public ingress, edge security, public DNS, edge TLS termination.

**Put here if:** It faces the public internet (gateway or load balancer with WAF, public DNS, edge TLS).

**Not here:**

- Core VNet routing → core-network
- Private data plane → data-platform

**Examples:**

- Application gateway or regional LB with WAF
- Public DNS zones
- Internet-facing TLS termination

## Modules in use today

None yet (stub layer). Create with `make scaffold-module CLOUD=gcp MODULE=<name>`, then register as `network-edge/<name>` in `boilerplate-vars.yml`.

**Reads from:** core-network, app-platform

**Provides:** `tests/config/contracts/gcp.yaml` (layer `network-edge`)

**How to place a new service:** `infra/AGENTS.md#placement`