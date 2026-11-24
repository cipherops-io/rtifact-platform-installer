#!/usr/bin/env bash
# Download OpenTofu for one platform, verify it against the release SHA256SUMS, install <dest>/tofu.
# Usage: fetch-tofu.sh <version> <goos> <goarch> <dest-dir>
set -euo pipefail

if [ "$#" -ne 4 ]; then
  echo "usage: $0 <version> <goos> <goarch> <dest-dir>" >&2
  exit 2
fi
VERSION="$1"
GOOS="$2"
GOARCH="$3"
DEST="$4"

case "$GOOS/$GOARCH" in
  linux/amd64 | linux/arm64 | darwin/amd64 | darwin/arm64) ;;
  *)
    echo "unsupported platform $GOOS/$GOARCH (linux and darwin, amd64 and arm64)" >&2
    exit 2
    ;;
esac

BASE="${TOFU_DOWNLOAD_BASE:-https://github.com/opentofu/opentofu/releases/download}/v${VERSION}"
ZIP="tofu_${VERSION}_${GOOS}_${GOARCH}.zip"
SUMS="tofu_${VERSION}_SHA256SUMS"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

sha256() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | awk '{print $1}'
  else
    shasum -a 256 "$1" | awk '{print $1}'
  fi
}

echo "Downloading $BASE/$ZIP"
curl -fsSL --retry 3 "$BASE/$ZIP" -o "$TMP/$ZIP"
curl -fsSL --retry 3 "$BASE/$SUMS" -o "$TMP/$SUMS"

want="$(awk -v f="$ZIP" '$2 == f || $2 == "*"f {print $1}' "$TMP/$SUMS")"
got="$(sha256 "$TMP/$ZIP")"
if [ -z "$want" ]; then
  echo "$ZIP is not listed in $SUMS" >&2
  exit 1
fi
if [ "$want" != "$got" ]; then
  echo "checksum mismatch for $ZIP: want $want, got $got" >&2
  exit 1
fi

unzip -q "$TMP/$ZIP" tofu -d "$TMP/out"
mkdir -p "$DEST"
install -m 0755 "$TMP/out/tofu" "$DEST/tofu"
echo "Installed $DEST/tofu (OpenTofu $VERSION, $GOOS/$GOARCH, sha256 verified)"
