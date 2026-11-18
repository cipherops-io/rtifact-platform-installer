#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
MIRROR="$ROOT/build/embed/mirror"
mkdir -p "$MIRROR"

pick_tofu() {
  if [ -n "${TOFU_BIN:-}" ] && command -v "$TOFU_BIN" >/dev/null 2>&1; then
    echo "$TOFU_BIN"
    return
  fi
  if command -v tofu >/dev/null 2>&1; then
    command -v tofu
    return
  fi
  HOST_GOOS="$(go env GOOS)"
  HOST_GOARCH="$(go env GOARCH)"
  CAND="$ROOT/build/embed/tofu/${HOST_GOOS}_${HOST_GOARCH}/tofu"
  if [ -x "$CAND" ]; then
    echo "$CAND"
    return
  fi
  echo "tofu not found; install OpenTofu or run fetch-tofu.sh for this host" >&2
  exit 1
}

TOFU="$(pick_tofu)"

WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT
cp -R "$ROOT/infra" "$WORK/infra"

cd "$WORK/infra/stacks/ai-os-aks"
"$TOFU" init -input=false
"$TOFU" providers mirror "$MIRROR"
echo "Provider mirror written to $MIRROR"