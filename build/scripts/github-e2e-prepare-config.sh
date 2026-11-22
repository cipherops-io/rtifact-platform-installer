#!/usr/bin/env bash
# Substitute Azure IDs into examples/azure/github-e2e-azure.yaml for CI (writes to a temp file).
set -euo pipefail
SRC="${1:-examples/azure/github-e2e-azure.yaml}"
OUT="${2:-/tmp/github-e2e-azure.yaml}"
: "${AZURE_SUBSCRIPTION_ID:?AZURE_SUBSCRIPTION_ID required}"
: "${AZURE_TENANT_ID:?AZURE_TENANT_ID required}"
sed \
  -e "s/SUBSCRIPTION_ID_PLACEHOLDER/${AZURE_SUBSCRIPTION_ID}/g" \
  -e "s/TENANT_ID_PLACEHOLDER/${AZURE_TENANT_ID}/g" \
  "$SRC" > "$OUT"
echo "Wrote $OUT"