<!-- DO NOT EDIT — managed by cipherops-io/boilerplates (rtifact-product).
     Change it there, then run: make boilerplate-update -->

# AGENTS.md — core-network (azure)

## Purpose

Scope, naming, core networking, egress control, hybrid connectivity, jump host.

**Put here if:** It is foundation other layers build on (network, subnets, NAT, egress firewall, jump host, VPN/ExpressRoute).

**Not here:**

- Public ingress → network-edge
- Managed data services → data-platform
- Private endpoints for PaaS data → data-platform

**Examples:**

- VNet/VPC and subnets
- NAT and egress firewall
- Bastion / jump host
- VPN or ExpressRoute

## Modules in use today

- `naming`

- `network`

- `firewall`

- `bastion-host`

- `connectivity`


**Provides:** `docs/layer-contracts.md#core-network`

**How to place a new service:** `infra/AGENTS.md#placement`