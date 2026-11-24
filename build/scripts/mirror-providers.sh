#!/usr/bin/env bash
# Mirror the providers of every layer of each packaged cloud, for the target platform only.
# Env: TOFU_BIN (host OpenTofu, required), GOOS, GOARCH (target), PACKAGED_CLOUDS, MIRROR (output dir).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
: "${TOFU_BIN:?TOFU_BIN must point at a host-platform tofu (make host-tofu)}"
: "${GOOS:?GOOS (target) is required}"
: "${GOARCH:?GOARCH (target) is required}"
PACKAGED_CLOUDS="${PACKAGED_CLOUDS:-azure,aws,gcp}"
MIRROR="${MIRROR:-$ROOT/build/embed/mirror}"
PLATFORM="${GOOS}_${GOARCH}"

[ -x "$TOFU_BIN" ] || { echo "TOFU_BIN=$TOFU_BIN is not executable" >&2; exit 1; }
mkdir -p "$MIRROR"
MIRROR="$(cd "$MIRROR" && pwd)"

# `tofu providers mirror` downloads again even when the archive is present; skip layers already covered.
layer_mirrored() {
  local lock="$1/.terraform.lock.hcl" src ver
  [ -f "$lock" ] || return 1
  while read -r src ver; do
    [ -f "$MIRROR/$src/terraform-provider-${src##*/}_${ver}_${PLATFORM}.zip" ] || return 1
  done < <(awk '/^provider "/ {gsub(/"/, "", $2); src=$2} /^[[:space:]]*version[[:space:]]*=/ && src != "" {gsub(/"/, "", $3); print src, $3; src=""}' "$lock")
}

profile_stack_id() {
  awk -F: '/^[[:space:]]*stack_id:/{gsub(/^[[:space:]]+|[[:space:]]+$/, "", $2); gsub(/"/, "", $2); print $2; exit}' "$1"
}

WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT
mkdir -p "$WORK/infra/stacks"
[ -d "$ROOT/infra/modules" ] && cp -R "$ROOT/infra/modules" "$WORK/infra/modules"

IFS=',' read -ra CLOUDS <<< "$PACKAGED_CLOUDS"
layers=0
for profile_yaml in "$ROOT/profiles"/*/profile.yaml; do
  [ -f "$profile_yaml" ] || continue
  stack_id="$(profile_stack_id "$profile_yaml")"
  [ -n "$stack_id" ] || continue
  for c in "${CLOUDS[@]}"; do
    c="$(echo "$c" | xargs)"
    src="$ROOT/infra/stacks/$stack_id/$c"
    [ -n "$c" ] && [ -d "$src" ] || continue
    mkdir -p "$WORK/infra/stacks/$stack_id"
    rsync -a --exclude '.terraform' "$src" "$WORK/infra/stacks/$stack_id/"
    for layer in "$WORK/infra/stacks/$stack_id/$c"/*/; do
      layer="${layer%/}"
      [ -f "$layer/providers.tofu" ] || [ -f "$layer/versions.tofu" ] || continue
      layers=$((layers + 1))
      if layer_mirrored "$layer"; then
        echo "==> $stack_id/$c/$(basename "$layer"): providers already mirrored ($PLATFORM)"
        continue
      fi
      echo "==> mirror $stack_id/$c/$(basename "$layer") ($PLATFORM)"
      (cd "$layer" && "$TOFU_BIN" get >/dev/null && "$TOFU_BIN" providers mirror -platform="$PLATFORM" "$MIRROR")
    done
  done
done

if [ "$layers" -eq 0 ]; then
  echo "no layers found for clouds $PACKAGED_CLOUDS" >&2
  exit 1
fi
echo "Provider mirror written to $MIRROR ($layers layers, clouds: $PACKAGED_CLOUDS, platform: $PLATFORM)"
