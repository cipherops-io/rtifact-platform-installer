#!/bin/sh
# Install the rtifact binary for this machine from the Rtifact registry.
#
#   install.sh --cloud aws [--version v0.0.1] [--dir .] [--key cosign.pub]
#
# Uses oras when it is installed, otherwise curl against the registry API.
# Credentials: RTIFACT_REGISTRY_USERNAME / RTIFACT_REGISTRY_PASSWORD (or an existing `oras login`).
# Overrides: RTIFACT_REGISTRY (default zot.prod.rtifact.ai), RTIFACT_REPOSITORY (default rtifact-installer),
# RTIFACT_REGISTRY_PLAIN_HTTP=true for a registry without TLS.
set -eu

REGISTRY="${RTIFACT_REGISTRY:-zot.prod.rtifact.ai}"
REPOSITORY="${RTIFACT_REPOSITORY:-rtifact-installer}"
CLOUD=""
VERSION=""
DIR="."
KEY=""

die() { echo "install.sh: $*" >&2; exit 1; }

usage() {
  sed -n '2,8s/^# \{0,1\}//p' "$0"
  exit "${1:-0}"
}

while [ $# -gt 0 ]; do
  case "$1" in
    --cloud) CLOUD="${2:-}"; shift 2 ;;
    --version) VERSION="${2:-}"; shift 2 ;;
    --dir) DIR="${2:-}"; shift 2 ;;
    --key) KEY="${2:-}"; shift 2 ;;
    -h|--help) usage 0 ;;
    *) echo "install.sh: unknown argument: $1" >&2; usage 1 ;;
  esac
done

case "$CLOUD" in
  azure|aws|gcp) ;;
  "") die "--cloud is required (azure, aws or gcp)" ;;
  *) die "unknown cloud: $CLOUD (azure, aws or gcp)" ;;
esac

case "$(uname -s)" in
  Linux) OS=linux ;;
  Darwin) OS=darwin ;;
  *) die "unsupported OS: $(uname -s) (linux and darwin only)" ;;
esac
case "$(uname -m)" in
  x86_64|amd64) ARCH=amd64 ;;
  arm64|aarch64) ARCH=arm64 ;;
  *) die "unsupported architecture: $(uname -m) (amd64 and arm64 only)" ;;
esac

sha256() {
  if command -v sha256sum >/dev/null 2>&1; then sha256sum "$1" | awk '{print $1}'
  else shasum -a 256 "$1" | awk '{print $1}'; fi
}

WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT INT TERM
REF_BASE="$REGISTRY/$REPOSITORY"
ORAS_HTTP=""
COSIGN_HTTP=""
SCHEME=https
if [ "${RTIFACT_REGISTRY_PLAIN_HTTP:-}" = true ]; then
  ORAS_HTTP=--plain-http; COSIGN_HTTP="--allow-http-registry --allow-insecure-registry"; SCHEME=http
fi

# shellcheck disable=SC2086 # $ORAS_HTTP and $plat are empty or whole flags.
if command -v oras >/dev/null 2>&1; then
  set --
  if [ -n "${RTIFACT_REGISTRY_USERNAME:-}" ]; then
    set -- --username "$RTIFACT_REGISTRY_USERNAME" --password-stdin
  fi
  # latest-<cloud> is an index; a versioned tag names the platform artifact directly.
  if [ -n "$VERSION" ]; then ref="$REF_BASE:$VERSION-$CLOUD-$OS-$ARCH"; plat=""
  else ref="$REF_BASE:latest-$CLOUD"; plat="--platform $OS/$ARCH"; fi
  echo "Pulling $ref ($OS/$ARCH) with oras"
  printf '%s' "${RTIFACT_REGISTRY_PASSWORD:-}" |
    oras pull $ORAS_HTTP $plat -o "$WORK" "$@" "$ref" >/dev/null
  digest="$(printf '%s' "${RTIFACT_REGISTRY_PASSWORD:-}" |
    oras manifest fetch $ORAS_HTTP $plat "$@" "$ref" |
    tr -d '\n ' | sed -n 's/.*"layers":\[{[^}]*"digest":"sha256:\([0-9a-f]*\)".*/\1/p')"
  manifest_digest="$(printf '%s' "${RTIFACT_REGISTRY_PASSWORD:-}" |
    oras resolve $ORAS_HTTP $plat "$@" "$ref")"
else
  command -v curl >/dev/null 2>&1 || die "oras or curl is required"
  set -- -fsSL
  if [ -n "${RTIFACT_REGISTRY_USERNAME:-}" ]; then
    set -- "$@" -u "$RTIFACT_REGISTRY_USERNAME:${RTIFACT_REGISTRY_PASSWORD:-}"
  fi
  api="$SCHEME://$REGISTRY/v2/$REPOSITORY"
  if [ -z "$VERSION" ]; then
    VERSION="$(curl "$@" -H "Accept: application/vnd.oci.image.index.v1+json" "$api/manifests/latest-$CLOUD" |
      tr -d '\n ' | sed -n 's/.*"org.opencontainers.image.version":"\([^"]*\)".*/\1/p')"
    [ -n "$VERSION" ] || die "could not resolve latest-$CLOUD in $REF_BASE"
  fi
  tag="$VERSION-$CLOUD-$OS-$ARCH"
  echo "Downloading $REF_BASE:$tag with curl"
  curl "$@" -H "Accept: application/vnd.oci.image.manifest.v1+json" -D "$WORK/headers" \
    -o "$WORK/manifest.json" "$api/manifests/$tag" || die "no artifact $REF_BASE:$tag"
  digest="$(tr -d '\n ' < "$WORK/manifest.json" |
    sed -n 's/.*"layers":\[{[^}]*"digest":"sha256:\([0-9a-f]*\)".*/\1/p')"
  [ -n "$digest" ] || die "manifest for $tag has no layer"
  manifest_digest="$(tr -d '\r' < "$WORK/headers" | awk 'tolower($1)=="docker-content-digest:"{print $2}' | tail -1)"
  curl "$@" -o "$WORK/rtifact" "$api/blobs/sha256:$digest"
fi

[ -f "$WORK/rtifact" ] || die "artifact did not contain the rtifact binary"
[ -n "$digest" ] || die "could not read the layer digest"
actual="$(sha256 "$WORK/rtifact")"
[ "$actual" = "$digest" ] || die "checksum mismatch: got $actual, registry says $digest"

if [ -n "$KEY" ]; then
  command -v cosign >/dev/null 2>&1 || die "--key needs cosign"
  [ -n "${manifest_digest:-}" ] || die "could not read the manifest digest to verify"
  # shellcheck disable=SC2086 # $COSIGN_HTTP is empty or whole flags.
  cosign verify $COSIGN_HTTP --key "$KEY" --insecure-ignore-tlog=true "$REF_BASE@$manifest_digest" >/dev/null 2>&1 ||
    die "signature verification failed for $REF_BASE@$manifest_digest"
  echo "Signature verified ($REF_BASE@$manifest_digest)"
fi

mkdir -p "$DIR"
chmod +x "$WORK/rtifact"
mv "$WORK/rtifact" "$DIR/rtifact"
echo "Installed $DIR/rtifact (sha256 $actual). Check it with: $DIR/rtifact doctor"
