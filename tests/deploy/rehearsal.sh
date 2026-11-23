#!/usr/bin/env bash
# Operator journey on the built binary against the deploy-e2e kind cluster and Zot
# registry (env from build/scripts/deploy-e2e.sh).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
BIN="${RTIFACT_BIN:-$ROOT/dist/rtifact}"
if [ ! -x "$BIN" ]; then
  echo "rehearsal: $BIN missing (run: make build-dev or set RTIFACT_BIN)" >&2
  exit 1
fi
for v in DEPLOY_E2E_KUBECONFIG DEPLOY_E2E_REGISTRY_HOST DEPLOY_E2E_EXPECTED_SERVER RTIFACT_REGISTRY_USER RTIFACT_REGISTRY_PASSWORD; do
  if [ -z "${!v:-}" ]; then
    echo "rehearsal: $v not set (run via make deploy-e2e)" >&2
    exit 1
  fi
done
export INSTALLER_CONTENT_ROOT="$ROOT"
KCFG="$DEPLOY_E2E_KUBECONFIG"
WORK="$(mktemp -d /tmp/rtifact-rehearsal.XXXX)"
NS_A=rehearsal-a
DEPLOYMENT=e2e-rehearsal

step() { echo "==> rehearsal: $*"; }
fail() {
  echo "rehearsal FAILED: $*" >&2
  exit 1
}
rt() { "$BIN" deploy "$@" --kubeconfig "$KCFG"; }
# expect_code CODE CMD... runs CMD and fails unless it exits with CODE.
expect_code() {
  local want="$1"
  shift
  set +e
  "$@"
  local got=$?
  set -e
  [ "$got" -eq "$want" ] || fail "'$*' exited $got, want $want"
}

cleanup() {
  local ec=$?
  rt destroy --yes "$WORK/v1.yaml" >/dev/null 2>&1 || true
  helm uninstall slow -n "$NS_A" --kubeconfig "$KCFG" >/dev/null 2>&1 || true
  rm -rf "$WORK"
  return "$ec"
}
trap cleanup EXIT

sed -e "s|^name_prefix: .*|name_prefix: rehearsal|" \
  -e "s|^environment: .*|environment: rehearsal|" \
  -e "s|expected_server: .*|expected_server: \"$DEPLOY_E2E_EXPECTED_SERVER\"|" \
  -e "s|host: 127.0.0.1:5000|host: $DEPLOY_E2E_REGISTRY_HOST|" \
  -e "s|deploy-a|$NS_A|g" -e "s|deploy-b|rehearsal-b|g" \
  "$ROOT/tests/deploy/fixtures/e2e.yaml" >"$WORK/v1.yaml"
awk '/name: ok-a/{f=1} f&&/version:/{sub(/"1.0.0"/,"\"1.0.1\""); f=0} {print}' "$WORK/v1.yaml" >"$WORK/v2.yaml"
{ cat "$WORK/v1.yaml"; echo; } >"$WORK/slow.yaml"
cat >>"$WORK/slow.yaml" <<EOF
  - name: slow
    namespace: $NS_A
    chart: broken
    version: "1.0.0"
    crds: chart
    requires: [ok-a]
EOF
grep -q '"1.0.1"' "$WORK/v2.yaml" || fail "version bump fixture not generated"

"$BIN" version

step "validate"
"$BIN" deploy validate "$WORK/v1.yaml"
"$BIN" deploy validate "$ROOT/examples/deploy/minimal.yaml"

step "plan before first apply reports changes (exit 2) and JSON schema_version"
expect_code 2 rt plan --detailed-exitcode "$WORK/v1.yaml"
rt plan -o json "$WORK/v1.yaml" | grep -Eq '"schema_version": ?1' || fail "plan JSON missing schema_version"

step "invalid -o is rejected"
expect_code 1 rt plan -o yaml "$WORK/v1.yaml"

step "apply"
rt apply --yes "$WORK/v1.yaml"
expect_code 0 rt plan --detailed-exitcode "$WORK/v1.yaml"
rt status -o json "$WORK/v1.yaml" >"$WORK/status.json"
[ "$(grep -c '"state": "deployed"' "$WORK/status.json")" -eq 2 ] || fail "want 2 deployed releases: $(cat "$WORK/status.json")"

step "upgrade ok-a to 1.0.1"
rt apply --yes "$WORK/v2.yaml"
rt status "$WORK/v2.yaml" | grep -q "/ok-a  state=deployed .*deployed=1.0.1" || fail "ok-a not at 1.0.1"

step "rollback ok-a (healthy release returns to 1.0.0, stays installed)"
rt rollback --yes "$WORK/v2.yaml" ok-a
rt status "$WORK/v2.yaml" | grep -q "/ok-a  state=drifted .*wanted=1.0.1  deployed=1.0.0  status=deployed" ||
  fail "ok-a after rollback: $(rt status "$WORK/v2.yaml")"
expect_code 2 rt plan --detailed-exitcode "$WORK/v2.yaml"

step "SIGINT during a slow install exits 130 and releases the lock"
rt apply --yes "$WORK/slow.yaml" &
pid=$!
for _ in $(seq 1 180); do
  if kubectl --kubeconfig "$KCFG" get secret -n "$NS_A" -l owner=helm,name=slow,status=pending-install -o name 2>/dev/null | grep -q .; then
    break
  fi
  sleep 1
done
kill -INT "$pid"
set +e
wait "$pid"
code=$?
set -e
[ "$code" -eq 130 ] || fail "interrupted apply exited $code, want 130"
if kubectl --kubeconfig "$KCFG" get lease -n rtifact-system "rtifact-deploy-lock-$DEPLOYMENT" >/dev/null 2>&1; then
  fail "deploy lock still held after interrupt"
fi
helm uninstall slow -n "$NS_A" --kubeconfig "$KCFG" >/dev/null 2>&1 || true

step "destroy"
rt destroy --yes "$WORK/v1.yaml"
rt status -o json "$WORK/v1.yaml" >"$WORK/status.json"
[ "$(grep -c '"state": "missing"' "$WORK/status.json")" -eq 2 ] || fail "want 2 missing releases after destroy: $(cat "$WORK/status.json")"

echo "rehearsal OK"
