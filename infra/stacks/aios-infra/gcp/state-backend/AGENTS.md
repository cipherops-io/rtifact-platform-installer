<!-- DO NOT EDIT — managed by cipherops-io/boilerplates (rtifact-product).
     Change it there, then run: make boilerplate-update -->

# AGENTS.md — state-backend (gcp)

## Purpose

Remote state storage for OpenTofu (bootstrap may use local state).

**Put here if:** It stores this installer's own remote state backend.

**Not here:**

- Workload resources → data-platform or app-platform

**Examples:**

- Azure storage account for tfstate
- AWS S3 bucket for tfstate
- GCS bucket for tfstate

## Modules in use today

- `state-backend`


**Provides:** `docs/layer-contracts.md#state-backend`

**How to place a new service:** `infra/AGENTS.md#placement`