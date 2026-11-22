<!-- DO NOT EDIT — managed by cipherops-io/boilerplates (rtifact-product).
     Change it there, then run: make boilerplate-update -->

# AGENTS.md — network-edge (aws)

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

- `dns-public`

- `certificates`


**Reads from:** core-network, app-platform

**Provides:** `tests/config/contracts/aws.yaml` (layer `network-edge`)

**How to place a new service:** `infra/AGENTS.md#placement`