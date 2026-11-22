#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
PKG="$ROOT/internal/packaged"
PACKAGED_CLOUDS="${PACKAGED_CLOUDS:-azure,aws,gcp}"

IFS=',' read -ra CLOUDS <<< "$PACKAGED_CLOUDS"

rm -rf "$PKG/profiles" "$PKG/infra" "$PKG/schemas"
mkdir -p "$PKG"

PROFILE_ID="${PROFILE_ID:-aios-infra}"
mkdir -p "$PKG/profiles/$PROFILE_ID"
cp "$ROOT/profiles/$PROFILE_ID/profile.yaml" "$PKG/profiles/$PROFILE_ID/"
cp -f "$ROOT/profiles/$PROFILE_ID/README.md" "$PKG/profiles/$PROFILE_ID/" 2>/dev/null || true

for c in "${CLOUDS[@]}"; do
  c="$(echo "$c" | xargs)"
  [ -n "$c" ] || continue
  if [ -d "$ROOT/profiles/$PROFILE_ID/$c" ]; then
    mkdir -p "$PKG/profiles/$PROFILE_ID/$c"
    rsync -a "$ROOT/profiles/$PROFILE_ID/$c/" "$PKG/profiles/$PROFILE_ID/$c/"
  fi
done

mkdir -p "$PKG/schemas"
for c in "${CLOUDS[@]}"; do
  c="$(echo "$c" | xargs)"
  if [ -f "$ROOT/schemas/${c}.schema.json" ]; then
    cp "$ROOT/schemas/${c}.schema.json" "$PKG/schemas/"
  fi
done
if [ -f "$ROOT/schemas/common.schema.json" ]; then
  cp "$ROOT/schemas/common.schema.json" "$PKG/schemas/"
fi
if [ -f "$ROOT/schemas/deploy.schema.json" ]; then
  cp "$ROOT/schemas/deploy.schema.json" "$PKG/schemas/"
fi

mkdir -p "$PKG/infra/stacks/aios-infra" "$PKG/infra/modules"
for c in "${CLOUDS[@]}"; do
  c="$(echo "$c" | xargs)"
  [ -n "$c" ] || continue
  if [ -d "$ROOT/infra/stacks/aios-infra/$c" ]; then
    rsync -a --exclude '.terraform' "$ROOT/infra/stacks/aios-infra/$c" "$PKG/infra/stacks/aios-infra/"
  fi
  if [ -d "$ROOT/infra/modules/$c" ]; then
    rsync -a "$ROOT/infra/modules/$c" "$PKG/infra/modules/"
  fi
done

for c in "${CLOUDS[@]}"; do
  c="$(echo "$c" | xargs)"
  if [ -d "$ROOT/examples/$c" ]; then
    mkdir -p "$PKG/examples"
    rsync -a "$ROOT/examples/$c" "$PKG/examples/"
  fi
done
if [ -d "$ROOT/examples/deploy" ]; then
  mkdir -p "$PKG/examples"
  rsync -a "$ROOT/examples/deploy" "$PKG/examples/"
fi

if [ -d "$ROOT/build/embed" ]; then
  mkdir -p "$PKG/build"
  rsync -a "$ROOT/build/embed/" "$PKG/build/embed/"
  echo "Embedded build/embed (tofu + mirror) synced"
fi
touch "$PKG/.packaged"
echo "Packaged content synced to $PKG (clouds: $PACKAGED_CLOUDS)"