#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
PKG="$ROOT/internal/packaged"

rm -rf "$PKG/profiles" "$PKG/infra" "$PKG/schemas"
mkdir -p "$PKG"

rsync -a --exclude '.terraform' "$ROOT/profiles" "$ROOT/infra" "$ROOT/schemas" "$PKG/"
if [ -d "$ROOT/build/embed" ]; then
  mkdir -p "$PKG/build"
  rsync -a "$ROOT/build/embed/" "$PKG/build/embed/"
  echo "Embedded build/embed (tofu + mirror) synced"
fi
touch "$PKG/.packaged"
echo "Packaged content synced to $PKG"