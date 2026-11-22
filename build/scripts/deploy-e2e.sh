#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
LOG="/tmp/rtifact-deploy-e2e.log"
ZOT_LOG="/tmp/zot-deploy-e2e.log"

ZOT_VERSION="${ZOT_VERSION:-2.1.0}"
HELM_VERSION="${HELM_VERSION:-3.16.4}"
REG_PORT="${DEPLOY_E2E_REG_PORT:-5000}"
REG_HOST="127.0.0.1:${REG_PORT}"
CLUSTER_NAME="${DEPLOY_E2E_CLUSTER:-rtifact-deploy-e2e}"
KUBECONFIG="${DEPLOY_E2E_KUBECONFIG:-/tmp/${CLUSTER_NAME}-kubeconfig}"
REG_USER="${DEPLOY_E2E_REGISTRY_USER:-deploy}"
REG_PASS="${DEPLOY_E2E_REGISTRY_PASSWORD:-deploypass}"

cleanup() {
  local ec=$?
  docker rm -f rtifact-deploy-zot >/dev/null 2>&1 || true
  kind delete cluster --name "$CLUSTER_NAME" >/dev/null 2>&1 || true
  rm -f "${ZOT_CONFIG:-}" "${ZOT_HTPASSWD:-}" "${KIND_CONFIG:-}" 2>/dev/null || true
  if [ "$ec" -ne 0 ]; then
    echo "deploy-e2e FAILED (exit $ec); see $LOG"
  fi
  return "$ec"
}
trap cleanup EXIT

exec > >(tee -a "$LOG") 2>&1

export DEPLOY_E2E_KUBECONFIG="$KUBECONFIG"
export DEPLOY_E2E_REGISTRY_HOST="$REG_HOST"
export DEPLOY_E2E_REGISTRY_USER="$REG_USER"
export DEPLOY_E2E_REGISTRY_PASSWORD="$REG_PASS"
export RTIFACT_REGISTRY_USER="$REG_USER"
export RTIFACT_REGISTRY_PASSWORD="$REG_PASS"

command -v docker >/dev/null
command -v curl >/dev/null

if ! command -v kind >/dev/null; then
  echo "kind not found on PATH"
  exit 1
fi

install_helm() {
  if command -v helm >/dev/null && helm version --short 2>/dev/null | grep -q "v${HELM_VERSION}"; then
    return
  fi
  local tgz="/tmp/helm-v${HELM_VERSION}-linux-amd64.tar.gz"
  if [ "$(uname -s)" = "Darwin" ]; then
    tgz="/tmp/helm-v${HELM_VERSION}-darwin-amd64.tar.gz"
    curl -fsSL "https://get.helm.sh/helm-v${HELM_VERSION}-darwin-amd64.tar.gz" -o "$tgz"
  else
    curl -fsSL "https://get.helm.sh/helm-v${HELM_VERSION}-linux-amd64.tar.gz" -o "$tgz"
  fi
  tar -xzf "$tgz" -C /tmp
  export PATH="/tmp/linux-amd64:/tmp/darwin-amd64:$PATH"
  helm version --short
}
install_helm

ZOT_CONFIG="$(mktemp /tmp/zot-config.XXXX.json)"
ZOT_HTPASSWD="$(mktemp /tmp/zot-htpasswd.XXXX)"
if command -v htpasswd >/dev/null; then
  htpasswd -Bbn "$REG_USER" "$REG_PASS" >"$ZOT_HTPASSWD"
else
  docker run --rm --entrypoint htpasswd httpd:2 -Bbn "$REG_USER" "$REG_PASS" >"$ZOT_HTPASSWD"
fi
cat >"$ZOT_CONFIG" <<EOF
{
  "distSpecVersion": "1.1.1",
  "storage": { "rootDirectory": "/var/lib/registry" },
  "http": { "address": "0.0.0.0", "port": "5000" },
  "extensions": {
    "auth": {
      "htpasswd": { "path": "/etc/zot/htpasswd" }
    }
  }
}
EOF

echo "==> zot ${ZOT_VERSION} on :${REG_PORT} (authenticated)"
docker rm -f rtifact-deploy-zot >/dev/null 2>&1 || true
docker run -d --name rtifact-deploy-zot -p "${REG_PORT}:5000" \
  -v "$ZOT_CONFIG:/etc/zot/config.json:ro" \
  -v "$ZOT_HTPASSWD:/etc/zot/htpasswd:ro" \
  "ghcr.io/project-zot/zot:v${ZOT_VERSION}-minimal" >/dev/null

for i in $(seq 1 60); do
  if curl -sf -u "${REG_USER}:${REG_PASS}" "http://${REG_HOST}/v2/" >/dev/null; then
  echo "zot ready"
    break
  fi
  if [ "$i" -eq 60 ]; then
    echo "zot failed to become ready"
    docker logs rtifact-deploy-zot 2>"$ZOT_LOG" || true
    exit 1
  fi
  sleep 1
done

KIND_CONFIG="$(mktemp /tmp/kind-deploy-e2e.XXXX.yaml)"
cat >"$KIND_CONFIG" <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."${REG_HOST}"]
    endpoint = ["http://${REG_HOST}"]
  [plugins."io.containerd.grpc.v1.cri".registry.configs."${REG_HOST}".tls]
    insecure_skip_verify = true
EOF

echo "==> kind cluster ${CLUSTER_NAME}"
kind delete cluster --name "$CLUSTER_NAME" >/dev/null 2>&1 || true
kind create cluster --name "$CLUSTER_NAME" --config "$KIND_CONFIG"
export KIND_PID=1
kind get kubeconfig --name "$CLUSTER_NAME" >"$KUBECONFIG"
chmod 600 "$KUBECONFIG"

SERVER="$(kubectl --kubeconfig "$KUBECONFIG" config view --minify -o jsonpath='{.clusters[0].cluster.server}')"
export DEPLOY_E2E_EXPECTED_SERVER="$SERVER"

push_chart() {
  local src="$1"
  local ver="$2"
  local name
  name=$(sed -n 's/^name: //p' "${src}/Chart.yaml" | head -1)
  pkg="/tmp/${name}-${ver}.tgz"
  helm package "$src" --version "$ver" -d /tmp >/dev/null
  helm push "$pkg" "oci://${REG_HOST}/charts" --plain-http \
    --username "$REG_USER" --password "$REG_PASS"
}

echo "==> push test charts to oci://${REG_HOST}/charts"
for chart in ok-a ok-b crd broken stateful secretful crd-breaking; do
  src="${ROOT}/tests/deploy/charts/${chart}"
  ver=$(sed -n 's/^version: //p' "${src}/Chart.yaml" | head -1)
  push_chart "$src" "$ver"
done
push_chart "${ROOT}/tests/deploy/charts/ok-a" "1.0.1"
push_chart "${ROOT}/tests/deploy/charts/crd-2" "2.0.0"

cd "$ROOT"
go test -tags e2e -count=1 ./tests/deploy/...

if [ -x "${ROOT}/tests/deploy/rehearsal.sh" ]; then
  echo "==> rehearsal.sh"
  "${ROOT}/tests/deploy/rehearsal.sh"
fi

echo "deploy-e2e OK"