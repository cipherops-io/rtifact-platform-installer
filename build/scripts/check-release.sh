#!/usr/bin/env bash
# Fail when a release binary was built without OpenTofu or providers for its platform.
# Usage: check-release.sh <binary> <goos> <goarch> <packaged-clouds> <min-mb>
set -euo pipefail

if [ "$#" -ne 5 ]; then
  echo "usage: $0 <binary> <goos> <goarch> <packaged-clouds> <min-mb>" >&2
  exit 2
fi
BIN="$1"
PLATFORM="$2_$3"
PACKAGED_CLOUDS="$4"
MIN_MB="$5"

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
EMBED="${CHECK_EMBED_DIR:-$ROOT/internal/packaged/build/embed}"
STACKS="${CHECK_STACKS_DIR:-$ROOT/internal/packaged/infra/stacks}"
fail=0

if [ ! -x "$EMBED/tofu/$PLATFORM/tofu" ]; then
  echo "missing embedded OpenTofu: $EMBED/tofu/$PLATFORM/tofu" >&2
  fail=1
fi
for other in "$EMBED"/tofu/*/; do
  [ -d "$other" ] || continue
  if [ "$(basename "$other")" != "$PLATFORM" ]; then
    echo "embedded OpenTofu for another platform: $other" >&2
    fail=1
  fi
done

IFS=',' read -ra CLOUDS <<< "$PACKAGED_CLOUDS"
want="$(mktemp)"
trap 'rm -f "$want"' EXIT
for c in "${CLOUDS[@]}"; do
  c="$(echo "$c" | xargs)"
  [ -n "$c" ] || continue
  found=0
  for lock in "$STACKS"/*/"$c"/*/.terraform.lock.hcl; do
    [ -f "$lock" ] || continue
    found=1
    awk '/^provider "/ {gsub(/"/, "", $2); src=$2} /^[[:space:]]*version[[:space:]]*=/ && src != "" {gsub(/"/, "", $3); print src, $3; src=""}' "$lock" >> "$want"
  done
  if [ "$found" -eq 0 ]; then
    echo "no packaged lock files for cloud $c under $STACKS" >&2
    fail=1
  fi
done

while read -r src ver; do
  name="${src##*/}"
  zip="$EMBED/mirror/$src/terraform-provider-${name}_${ver}_${PLATFORM}.zip"
  if [ ! -f "$zip" ]; then
    echo "missing provider in mirror: $src $ver ($PLATFORM)" >&2
    fail=1
  fi
done < <(sort -u "$want")

if [ -f "$BIN" ]; then
  size=$(wc -c < "$BIN" | tr -d ' ')
  min=$((MIN_MB * 1024 * 1024))
  echo "$BIN: $((size / 1024 / 1024)) MiB (floor ${MIN_MB} MiB)"
  if [ "$size" -lt "$min" ]; then
    echo "binary is smaller than ${MIN_MB} MiB; OpenTofu or providers were not embedded" >&2
    fail=1
  fi
else
  echo "binary not found: $BIN" >&2
  fail=1
fi

[ "$fail" -eq 0 ] && echo "release contents OK ($PLATFORM, clouds: $PACKAGED_CLOUDS)"
exit "$fail"
