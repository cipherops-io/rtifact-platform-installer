#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
PKG="$ROOT/internal/packaged"
PACKAGED_CLOUDS="${PACKAGED_CLOUDS:-azure,aws,gcp}"

IFS=',' read -ra CLOUDS <<< "$PACKAGED_CLOUDS"

rm -rf "$PKG/profiles" "$PKG/infra" "$PKG/schemas"
mkdir -p "$PKG"

rsync -a "$ROOT/profiles/aios-infra" "$PKG/profiles/"
rsync -a "$ROOT/schemas" "$PKG/"
mkdir -p "$PKG/infra/stacks/aios-infra" "$PKG/infra/modules"
for c in "${CLOUDS[@]}"; do
  c="${c#"${c%%[![:space:]]*}"}"
  c="${c%"${c##*[![:space:]]}"}"
  [ -n "$c" ] || continue
  if [ -d "$ROOT/infra/stacks/aios-infra/$c" ]; then
    rsync -a "$ROOT/infra/stacks/aios-infra/$c" "$PKG/infra/stacks/aios-infra/"
  fi
  if [ -d "$ROOT/infra/modules/$c" ]; then
    rsync -a "$ROOT/infra/modules/$c" "$PKG/infra/modules/"
  fi
done

if [ -d "$ROOT/build/embed" ]; then
  mkdir -p "$PKG/build"
  rsync -a "$ROOT/build/embed/" "$PKG/build/embed/"
  echo "Embedded build/embed (tofu + mirror) synced"
fi
touch "$PKG/.packaged"
echo "Packaged content synced to $PKG (clouds: $PACKAGED_CLOUDS)"