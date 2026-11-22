<!-- DO NOT EDIT — managed by cipherops-io/boilerplates (rtifact-product).
     Change it there, then run: make boilerplate-update -->

# AGENTS.md — ai-integrations (azure)

## Purpose

Managed AI model access and transactional email / communications.

**Put here if:** It is managed AI model access or transactional email.

**Not here:**

- General data stores for AI → data-platform (keys, storage for Foundry live in data-platform)

**Examples:**

- Azure AI Foundry hub
- Communication / email services

## Modules in use today

- `ai-foundry`


**Reads from:** core-network, data-platform

**Provides:** `tests/config/contracts/azure.yaml` (layer `ai-integrations`)

**How to place a new service:** `infra/AGENTS.md#placement`