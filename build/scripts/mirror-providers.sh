#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
MIRROR="$ROOT/build/embed/mirror"
mkdir -p "$MIRROR"
PACKAGED_CLOUDS="${PACKAGED_CLOUDS:-azure,aws,gcp}"

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
mkdir -p "$WORK/infra/stacks/aios-infra"
cp -R "$ROOT/infra/modules" "$WORK/infra/" 2>/dev/null || true

IFS=',' read -ra CLOUDS <<< "$PACKAGED_CLOUDS"
for c in "${CLOUDS[@]}"; do
  c="${c#"${c%%[![:space:]]*}"}"
  c="${c%"${c##*[![:space:]]}"}"
  [ -n "$c" ] || continue
  src="$ROOT/infra/stacks/aios-infra/$c"
  [ -d "$src" ] || continue
  cp -R "$src" "$WORK/infra/stacks/aios-infra/"
  layer="$WORK/infra/stacks/aios-infra/$c/core-network"
  if [ -f "$layer/providers.tofu" ]; then
    (cd "$layer" && "$TOFU" init -backend=false -input=false && "$TOFU" providers mirror "$MIRROR")
  fi
done

echo "Provider mirror written to $MIRROR (clouds: $PACKAGED_CLOUDS)"