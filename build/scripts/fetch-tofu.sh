#!/usr/bin/env bash
set -euo pipefail

VERSION="${1:-1.9.0}"
GOOS="${2:-$(go env GOOS)}"
GOARCH="${3:-$(go env GOARCH)}"

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
DEST="$ROOT/build/embed/tofu/${GOOS}_${GOARCH}"
mkdir -p "$DEST"

ZIP="tofu_${VERSION}_${GOOS}_${GOARCH}.zip"
URL="https://github.com/opentofu/opentofu/releases/download/v${VERSION}/${ZIP}"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

echo "Downloading $URL"
curl -fsSL "$URL" -o "$TMP/$ZIP"
unzip -q "$TMP/$ZIP" -d "$TMP"

BIN="$DEST/tofu"
if [ "$GOOS" = "windows" ]; then
  BIN="$DEST/tofu.exe"
fi
install -m 0755 "$TMP/tofu${GOOS:+.exe}" "$BIN" 2>/dev/null || install -m 0755 "$TMP/tofu" "$BIN"
echo "Installed $BIN"