<!-- DO NOT EDIT — managed by cipherops-io/boilerplates (rtifact-product).
     Change it there, then run: make boilerplate-update -->

# AGENTS.md — ai-integrations (gcp)

## Purpose

Managed AI model access and transactional email / communications.

**Put here if:** It is managed AI model access or transactional email.

**Not here:**

- General data stores for AI → data-platform (keys, storage for Foundry live in data-platform)

**Examples:**

- Azure AI Foundry hub
- Communication / email services

## Modules in use today

None yet (stub layer). Create with `make scaffold-module CLOUD=gcp MODULE=<name>`, then register as `ai-integrations/<name>` in `boilerplate-vars.yml`.

**Reads from:** core-network, data-platform

**Provides:** `tests/config/contracts/gcp.yaml` (layer `ai-integrations`)

**How to place a new service:** `infra/AGENTS.md#placement`