<!-- DO NOT EDIT — managed by cipherops-io/boilerplates (rtifact-product).
     Change it there, then run: make boilerplate-update -->

# AGENTS.md — app-platform (aws)

## Purpose

Kubernetes clusters, node pools, workload identity, compute VMs.

**Put here if:** It runs workloads (Kubernetes, node pools, VMs, workload identity).

**Not here:**

- Databases or key vaults → data-platform
- Public edge TLS → network-edge

**Examples:**

- AKS / EKS / GKE
- Node pools and workload identity
- Standalone compute VMs

## Modules in use today

- `eks`

- `compute`

- `karpenter-infra`


**Reads from:** core-network, data-platform

**Provides:** `tests/config/contracts/aws.yaml` (layer `app-platform`)

**How to place a new service:** `infra/AGENTS.md#placement`