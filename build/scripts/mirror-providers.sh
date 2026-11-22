#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
MIRROR="$ROOT/build/embed/mirror"
mkdir -p "$MIRROR"
PACKAGED_CLOUDS="${PACKAGED_CLOUDS:-azure,aws,gcp}"

profile_stack_id() {
  awk -F: '/^[[:space:]]*stack_id:/{gsub(/^[[:space:]]+|[[:space:]]+$/, "", $2); gsub(/"/, "", $2); print $2; exit}' "$1"
}

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
cp -R "$ROOT/infra/modules" "$WORK/infra/" 2>/dev/null || true

IFS=',' read -ra CLOUDS <<< "$PACKAGED_CLOUDS"

for profile_yaml in "$ROOT/profiles"/*/profile.yaml; do
  [ -f "$profile_yaml" ] || continue
  stack_id="$(profile_stack_id "$profile_yaml")"
  [ -n "$stack_id" ] || continue
  mkdir -p "$WORK/infra/stacks/$stack_id"
  for c in "${CLOUDS[@]}"; do
    c="$(echo "$c" | xargs)"
    [ -n "$c" ] || continue
    src="$ROOT/infra/stacks/$stack_id/$c"
    [ -d "$src" ] || continue
    cp -R "$src" "$WORK/infra/stacks/$stack_id/"
    layer="$WORK/infra/stacks/$stack_id/$c/core-network"
    if [ -f "$layer/providers.tofu" ]; then
      (cd "$layer" && "$TOFU" init -backend=false -input=false && "$TOFU" providers mirror "$MIRROR")
    fi
  done
done

echo "Provider mirror written to $MIRROR (clouds: $PACKAGED_CLOUDS)"