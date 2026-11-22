#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
PKG="$ROOT/internal/packaged"
PACKAGED_CLOUDS="${PACKAGED_CLOUDS:-azure,aws,gcp}"

IFS=',' read -ra CLOUDS <<< "$PACKAGED_CLOUDS"

profile_stack_id() {
  awk -F: '/^[[:space:]]*stack_id:/{gsub(/^[[:space:]]+|[[:space:]]+$/, "", $2); gsub(/"/, "", $2); print $2; exit}' "$1"
}

rm -rf "$PKG/profiles" "$PKG/infra" "$PKG/schemas"
mkdir -p "$PKG"

for profile_dir in "$ROOT/profiles"/*/; do
  [ -d "$profile_dir" ] || continue
  profile_yaml="$profile_dir/profile.yaml"
  [ -f "$profile_yaml" ] || continue
  profile_id="$(basename "$profile_dir")"
  stack_id="$(profile_stack_id "$profile_yaml")"
  [ -n "$stack_id" ] || { echo "profile $profile_id: missing stack_id in profile.yaml" >&2; exit 1; }

  mkdir -p "$PKG/profiles/$profile_id"
  cp "$profile_yaml" "$PKG/profiles/$profile_id/"
  cp -f "$profile_dir/README.md" "$PKG/profiles/$profile_id/" 2>/dev/null || true
  if [ -f "$profile_dir/checklist.md" ]; then
    cp "$profile_dir/checklist.md" "$PKG/profiles/$profile_id/"
  fi

  for c in "${CLOUDS[@]}"; do
    c="$(echo "$c" | xargs)"
    [ -n "$c" ] || continue
    if [ -d "$profile_dir/$c" ]; then
      mkdir -p "$PKG/profiles/$profile_id/$c"
      rsync -a "$profile_dir/$c/" "$PKG/profiles/$profile_id/$c/"
    fi
  done

  mkdir -p "$PKG/infra/stacks/$stack_id" "$PKG/infra/modules"
  for c in "${CLOUDS[@]}"; do
    c="$(echo "$c" | xargs)"
    [ -n "$c" ] || continue
    if [ -d "$ROOT/infra/stacks/$stack_id/$c" ]; then
      rsync -a --exclude '.terraform' "$ROOT/infra/stacks/$stack_id/$c" "$PKG/infra/stacks/$stack_id/"
    fi
  done
done

if [ -d "$ROOT/infra/modules" ]; then
  mkdir -p "$PKG/infra/modules"
  for c in "${CLOUDS[@]}"; do
    c="$(echo "$c" | xargs)"
    if [ -d "$ROOT/infra/modules/$c" ]; then
      rsync -a "$ROOT/infra/modules/$c" "$PKG/infra/modules/"
    fi
  done
fi

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