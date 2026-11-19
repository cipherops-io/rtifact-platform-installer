# Azure OIDC + GitHub secrets — executed setup (Commotion sandbox)

This runbook records the **exact commands** used to wire GitHub Actions (`azure-e2e` workflow) to Azure subscription **cmtn-test** via OIDC (no client secret).

**Date executed:** 2026-11-19  
**Repository:** `cipherops-io/rtifact-platform-installer`  
**Entra app display name:** `github-rtifact-e2e-cipherops`

## Resulting IDs

| Item | Value |
|------|--------|
| Subscription (`AZURE_SUBSCRIPTION_ID`) | `0778eb6e-3c52-42b2-a0b2-89dbd4a176e0` |
| Tenant (`AZURE_TENANT_ID`) | `ad5a1011-36aa-4e9b-9374-121f94a45916` |
| App client ID (`AZURE_CLIENT_ID`) | `d6761651-136e-487d-9215-d5b06ad3b5cd` |
| App object ID (for federated creds) | `66f443c2-1a98-4595-aa1c-747eb74cf0da` |

Federated credential subjects:

- `repo:cipherops-io/rtifact-platform-installer:ref:refs/heads/main`
- `repo:cipherops-io/rtifact-platform-installer:environment:azure-e2e`

RBAC: **Contributor** on `/subscriptions/0778eb6e-3c52-42b2-a0b2-89dbd4a176e0`.

## Prerequisites

```bash
az login
az account set --subscription "0778eb6e-3c52-42b2-a0b2-89dbd4a176e0"
gh auth status   # must have admin on cipherops-io/rtifact-platform-installer
```

## Azure CLI — app registration, OIDC, RBAC

```bash
APP_NAME="github-rtifact-e2e-cipherops"
SUB_ID="0778eb6e-3c52-42b2-a0b2-89dbd4a176e0"
TENANT_ID="$(az account show --query tenantId -o tsv)"

# Create app (skip if it already exists)
APP_ID=$(az ad app list --display-name "$APP_NAME" --query "[0].appId" -o tsv)
if [ -z "$APP_ID" ] || [ "$APP_ID" = "null" ]; then
  APP_ID=$(az ad app create --display-name "$APP_NAME" --query appId -o tsv)
fi

OBJECT_ID=$(az ad app show --id "$APP_ID" --query id -o tsv)

# Service principal
SP_ID=$(az ad sp list --filter "appId eq '$APP_ID'" --query "[0].id" -o tsv)
if [ -z "$SP_ID" ] || [ "$SP_ID" = "null" ]; then
  az ad sp create --id "$APP_ID"
fi

# Federated credentials (GitHub Actions issuer)
for subj in \
  "repo:cipherops-io/rtifact-platform-installer:ref:refs/heads/main" \
  "repo:cipherops-io/rtifact-platform-installer:environment:azure-e2e"
do
  NAME=$(echo "$subj" | sed 's/[^a-zA-Z0-9]/-/g' | cut -c1-120)
  az ad app federated-credential create --id "$OBJECT_ID" --parameters "{
    \"name\": \"$NAME\",
    \"issuer\": \"https://token.actions.githubusercontent.com\",
    \"subject\": \"$subj\",
    \"audiences\": [\"api://AzureADTokenExchange\"]
  }" || true
done

# Contributor on subscription (E2E sandbox only)
az role assignment create \
  --assignee "$APP_ID" \
  --role "Contributor" \
  --scope "/subscriptions/$SUB_ID" \
  || true
```

### Verify

```bash
APP_ID="d6761651-136e-487d-9215-d5b06ad3b5cd"
OBJECT_ID="66f443c2-1a98-4595-aa1c-747eb74cf0da"
SUB_ID="0778eb6e-3c52-42b2-a0b2-89dbd4a176e0"

az ad app federated-credential list --id "$OBJECT_ID" -o table
az role assignment list --assignee "$APP_ID" --all \
  --query "[?contains(scope,'$SUB_ID')].{role:roleDefinitionName,scope:scope}" -o table
```

## GitHub CLI — secrets + environment

```bash
REPO="cipherops-io/rtifact-platform-installer"
APP_ID="d6761651-136e-487d-9215-d5b06ad3b5cd"
TENANT_ID="ad5a1011-36aa-4e9b-9374-121f94a45916"
SUB_ID="0778eb6e-3c52-42b2-a0b2-89dbd4a176e0"

gh secret set AZURE_CLIENT_ID --body "$APP_ID" --repo "$REPO"
gh secret set AZURE_TENANT_ID --body "$TENANT_ID" --repo "$REPO"
gh secret set AZURE_SUBSCRIPTION_ID --body "$SUB_ID" --repo "$REPO"

gh api -X PUT "repos/$REPO/environments/azure-e2e" --input - <<'EOF'
{"wait_timer":0}
EOF

gh secret list --repo "$REPO"
```

## Workflow action pins (latest as of 2026-11-19)

Fetched via `gh api repos/<owner>/<repo>/releases/latest`:

| Action | Tag |
|--------|-----|
| `actions/checkout` | `v7.0.1` |
| `actions/setup-go` | `v7.1.0` |
| `opentofu/setup-opentofu` | `v2.0.2` |
| `actions/upload-artifact` | `v7.0.2` |
| `azure/login` | `v3` (latest release `v3.1.0`) |

See [azure-github-oidc-e2e.md](./azure-github-oidc-e2e.md) for how to run **Actions → azure-e2e**.