#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
LOG="/tmp/rtifact-deploy-e2e.log"
exec > >(tee -a "$LOG") 2>&1

ZOT_VERSION="${ZOT_VERSION:-2.1.0}"
REG_PORT="${DEPLOY_E2E_REG_PORT:-5000}"
REG_HOST="127.0.0.1:${REG_PORT}"
CLUSTER_NAME="${DEPLOY_E2E_CLUSTER:-rtifact-deploy-e2e}"
KUBECONFIG="${DEPLOY_E2E_KUBECONFIG:-/tmp/${CLUSTER_NAME}-kubeconfig}"

export DEPLOY_E2E_KUBECONFIG="$KUBECONFIG"
export DEPLOY_E2E_REGISTRY_HOST="$REG_HOST"
export DEPLOY_E2E_REGISTRY_USER="${DEPLOY_E2E_REGISTRY_USER:-deploy}"
export DEPLOY_E2E_REGISTRY_PASSWORD="${DEPLOY_E2E_REGISTRY_PASSWORD:-deploypass}"
export RTIFACT_REGISTRY_USER="$DEPLOY_E2E_REGISTRY_USER"
export RTIFACT_REGISTRY_PASSWORD="$DEPLOY_E2E_REGISTRY_PASSWORD"

command -v docker >/dev/null
command -v kind >/dev/null
command -v helm >/dev/null

echo "==> zot ${ZOT_VERSION} on :${REG_PORT}"
docker rm -f rtifact-deploy-zot >/dev/null 2>&1 || true
docker run -d --name rtifact-deploy-zot -p "${REG_PORT}:5000" \
  "ghcr.io/project-zot/zot:v${ZOT_VERSION}-minimal" >/dev/null
sleep 2

REGISTRY_CONFIG="/tmp/kind-registries.yaml"
cat >"$REGISTRY_CONFIG" <<EOF
containerdRegistryMirrors:
- containerdPort: 5001
  endpoints:
  - http://${REG_HOST}
EOF

echo "==> kind cluster ${CLUSTER_NAME}"
kind delete cluster --name "$CLUSTER_NAME" >/dev/null 2>&1 || true
kind create cluster --name "$CLUSTER_NAME" --config "$REGISTRY_CONFIG"
kind get kubeconfig --name "$CLUSTER_NAME" >"$KUBECONFIG"
chmod 600 "$KUBECONFIG"

SERVER="$(kubectl --kubeconfig "$KUBECONFIG" config view --minify -o jsonpath='{.clusters[0].cluster.server}')"
export DEPLOY_E2E_EXPECTED_SERVER="$SERVER"

echo "==> push test charts to oci://${REG_HOST}/charts"
for chart in ok-a ok-b crd broken stateful secretful crd-breaking; do
  src="${ROOT}/tests/deploy/charts/${chart}"
  ver=$(sed -n 's/^version: //p' "${src}/Chart.yaml" | head -1)
  pkg="/tmp/${chart}-${ver}.tgz"
  helm package "$src" --version "$ver" -d /tmp >/dev/null
  helm push "$pkg" "oci://${REG_HOST}/charts" --plain-http
done
helm package "${ROOT}/tests/deploy/charts/ok-a" --version 1.0.1 -d /tmp >/dev/null
helm push /tmp/ok-a-1.0.1.tgz "oci://${REG_HOST}/charts" --plain-http
helm package "${ROOT}/tests/deploy/charts/crd" --version 2.0.0 -d /tmp >/dev/null
helm push /tmp/crd-2.0.0.tgz "oci://${REG_HOST}/charts" --plain-http

cd "$ROOT"
go test -tags e2e -count=1 ./tests/deploy/...

if [ -x "${ROOT}/tests/deploy/rehearsal.sh" ]; then
  echo "==> rehearsal.sh"
  "${ROOT}/tests/deploy/rehearsal.sh" || true
fi

echo "deploy-e2e OK"