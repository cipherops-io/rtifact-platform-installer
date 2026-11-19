# Azure E2E from GitHub Actions (OIDC)

This repo can **plan**, **apply**, and **destroy** Azure stacks under `infra/stacks/aios-infra/azure/*` using the **`rtifact`** CLI in GitHub Actions, with **no client secret**—only short-lived tokens via **OpenID Connect**.

AWS/GCP deploy automation is **not enabled** in Actions yet; smoke tests remain in `tests/e2e/aws_test.go` and `tests/e2e/gcp_test.go`.

## What runs where

| Layer | Path | Typical rtifact command |
|-------|------|-------------------------|
| state-backend | `.../state-backend` | `bootstrap` or `--stack state-backend` |
| core-network | `.../core-network` | `plan` / `apply` / `destroy` |
| data-platform | `.../data-platform` | same |
| app-platform | `.../app-platform` | same |
| network-edge | `.../network-edge` | same |
| ai-integrations | `.../ai-integrations` | same |

Apply order for `--stack all`: `core-network` → `data-platform` → `app-platform` → `network-edge` → `ai-integrations` (state-backend via bootstrap when using remote state). Destroy reverses that order.

Default E2E config: [`examples/github-e2e-azure.yaml`](../examples/github-e2e-azure.yaml) (Commotion-shaped sandbox with isolated `10.21.0.0/20` CIDRs and `gh-rtifact-e2e-sbx` prefix). Subscription and tenant IDs are injected in CI via `scripts/github-e2e-prepare-config.sh`.

Workflow **azure-e2e** (`workflow_dispatch` actions): **validate**, **plan**, **apply** (bootstrap + state-backend + stack), **bootstrap-apply-destroy** (full lifecycle), **destroy-all** (workloads then state-backend).

## Test locally (same as Actions)

```bash
cd rtifact-platform-installer
az login
export ARM_USE_OIDC=false  # use Azure CLI session locally
export INSTALLER_CONTENT_ROOT=.
make build-dev PACKAGED_CLOUDS=azure

export E2E_AZURE_CONFIG=../commotion-ai-os-sandbox.yaml  # or your config
export E2E_AZURE_ACTION=plan   # plan | apply | destroy
export E2E_AZURE_STACK=core-network   # or all

go test -tags=e2e -timeout=180m ./tests/e2e/ -run TestAzureStackAction -count=1
```

Smoke only (no plan/apply):

```bash
export E2E_AZURE_CONFIG=examples/sandbox.yaml
make e2e-azure-smoke
```

## One-time Azure OIDC setup

**Automated runbook (commands already executed for Commotion sandbox):** [azure-github-oidc-setup-commands.md](./azure-github-oidc-setup-commands.md)

### Manual steps (if reproducing elsewhere)

### 1. App registration (Entra ID)

1. Entra ID → **App registrations** → **New registration**
2. Name: `github-rtifact-e2e` (or similar)
3. Supported account types: single tenant
4. Note **Application (client) ID** and **Directory (tenant) ID**

### 2. Federated credential (GitHub → Entra)

1. App → **Certificates & secrets** → **Federated credentials** → **Add credential**
2. **Federated credential scenario**: GitHub Actions deploying Azure resources
3. Organization: `cipherops-io`
4. Repository: `rtifact-platform-installer`
5. Entity type: **Branch** → `main` (or **Environment** → `azure-e2e` if you use the GitHub Environment)
6. Name: `github-main` (or `github-azure-e2e`)

Repeat for `refs/heads/main` and optionally `environment:azure-e2e` if you protect E2E with a GitHub Environment.

**Environment jobs:** GitHub may emit a subject with numeric org/repo IDs, e.g.  
`repo:cipherops-io@167057720/rtifact-platform-installer@1463839531:environment:azure-e2e`  
—not the string form `repo:org/repo:environment:...`. Add a **second** federated credential whose subject matches the `subject claim` line in the `azure/login` log ([runbook](./azure-github-oidc-setup-commands.md)).

### 3. RBAC on subscription

Assign the app **Contributor** (or a custom role) on the **target subscription** used for E2E (sandbox only):

```bash
APP_ID="<application-client-id>"
SUB_ID="<subscription-id>"
az role assignment create \
  --assignee "$APP_ID" \
  --role Contributor \
  --subscription "$SUB_ID"
```

For least privilege, scope to a dedicated resource group instead of the whole subscription.

### 4. GitHub repository secrets

In **Settings → Secrets and variables → Actions**, add:

| Secret | Value |
|--------|--------|
| `AZURE_CLIENT_ID` | App registration client ID |
| `AZURE_TENANT_ID` | Tenant ID |
| `AZURE_SUBSCRIPTION_ID` | E2E subscription ID |

No `AZURE_CLIENT_SECRET` is required for OIDC.

### 5. GitHub Environment (recommended)

1. **Settings → Environments** → create **`azure-e2e`**
2. Optional: required reviewers, deployment branch rule `main`
3. The workflow `azure-e2e.yml` targets `environment: azure-e2e`

### 6. Run the workflow

**Actions → azure-e2e → Run workflow**

Defaults: **plan** (after **validate** with Azure quota check on), **stack = all**. This workflow does **not** run apply or destroy.

On a greenfield subscription, **`plan --stack all`** plans **core-network** only; layers that depend on upstream OpenTofu state are **skipped** until those upstream stacks have been applied at least once.

1. Use defaults or set **validate** only to skip plan
2. Toggle **check_azure** off if you need validate without `az` quota API

## OpenTofu auth in the runner

After `azure/login@v3`, the workflow sets:

- `ARM_USE_OIDC=true`
- `ARM_CLIENT_ID`, `ARM_TENANT_ID`, `ARM_SUBSCRIPTION_ID`

The embedded stacks use the `azurerm` provider; these variables match [Azure Provider workload identity / OIDC](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/guides/service_principal_oidc) for GitHub Actions.

## When you are done with OIDC setup

Reply that secrets and federated credentials are in place. We can trigger **azure-e2e** with `plan` on `core-network`, then proceed to `apply`/`destroy` as needed.