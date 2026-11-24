#!/usr/bin/env bash
# Publish release binaries to an OCI registry: one artifact per cloud and platform, one index per cloud,
# a SHA256SUMS artifact, cosign signatures, and latest-<cloud> for non-prerelease versions.
# Exact-version tags are immutable: the same content is skipped, other content fails unless REPLACE_EXISTING=true.
#
# Env (required): DIST_DIR (holds <product>-<cloud>-<goos>-<goarch>), REGISTRY, REPOSITORY, VERSION, CLOUDS, PLATFORMS.
# Env (optional): PRODUCT (rtifact), REPLACE_EXISTING (false), SIGN (true; needs COSIGN_PRIVATE_KEY, COSIGN_PASSWORD),
#   COSIGN_PUBLIC_KEY_FILE (build/cosign.pub), PLAIN_HTTP (false; local test registries),
#   GIT_COMMIT, SOURCE_REPO, TOFU_VERSION (annotations), SUMMARY_FILE (markdown summary; default stdout).
# Runs on bash 3.2 (macOS) as well as bash 5.
set -euo pipefail

: "${DIST_DIR:?}" "${REGISTRY:?}" "${REPOSITORY:?}" "${VERSION:?}" "${CLOUDS:?}" "${PLATFORMS:?}"
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
PRODUCT="${PRODUCT:-rtifact}"
REPLACE_EXISTING="${REPLACE_EXISTING:-false}"
SIGN="${SIGN:-true}"
PUBKEY_FILE="${COSIGN_PUBLIC_KEY_FILE:-$ROOT/build/cosign.pub}"
SUMMARY_FILE="${SUMMARY_FILE:-/dev/stdout}"
ARTIFACT_TYPE="application/vnd.rtifact.installer.v1"
SUMS_TYPE="application/vnd.rtifact.installer.checksums.v1"
REPO="$REGISTRY/$REPOSITORY"

[[ "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.]+)?$ ]] || { echo "VERSION must look like v1.2.3 or v1.2.3-rc.1" >&2; exit 1; }
PRERELEASE=false
[[ "$VERSION" == *-* ]] && PRERELEASE=true

ORAS_FLAGS=""
COSIGN_REG=""
if [ "${PLAIN_HTTP:-false}" = true ]; then
  ORAS_FLAGS="--plain-http"
  COSIGN_REG="--allow-http-registry --allow-insecure-registry"
fi

WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT
mkdir -p "$WORK/kv"
kname() { printf '%s' "$1" | tr '|/:' '___'; }
kset() { printf '%s' "$2" > "$WORK/kv/$(kname "$1")"; }
kget() { cat "$WORK/kv/$(kname "$1")" 2>/dev/null || true; }

sha256() {
  if command -v sha256sum >/dev/null 2>&1; then sha256sum "$1" | awk '{print $1}'; else shasum -a 256 "$1" | awk '{print $1}'; fi
}

# resolve <tag>: prints the manifest digest, nothing when the tag is absent; fails on any other registry error.
resolve() {
  local out
  # shellcheck disable=SC2086
  if out=$(oras resolve $ORAS_FLAGS "$REPO:$1" 2>&1); then
    echo "$out"
  elif grep -qiE 'not ?found|manifest unknown|404' <<< "$out"; then
    return 0
  else
    echo "registry error resolving $REPO:$1: $out" >&2
    exit 1
  fi
}
# shellcheck disable=SC2086
layer_digest() { oras manifest fetch $ORAS_FLAGS "$REPO@$1" | jq -r '.layers[0].digest // empty'; }
# shellcheck disable=SC2086
index_members() { oras manifest fetch $ORAS_FLAGS "$REPO@$1" | jq -r '.manifests[].digest' | sort; }

IFS=',' read -ra CLOUD_LIST <<< "$CLOUDS"
IFS=',' read -ra PLATFORM_LIST <<< "$PLATFORMS"

# Inputs and SHA256SUMS
: > "$WORK/SHA256SUMS"
for c in "${CLOUD_LIST[@]}"; do
  for p in "${PLATFORM_LIST[@]}"; do
    name="$PRODUCT-$c-${p/\//-}"
    f="$DIST_DIR/$name"
    [ -f "$f" ] || { echo "missing $f" >&2; exit 1; }
    s="$(sha256 "$f")"
    kset "file|$c|$p" "$f"
    kset "sum|$c|$p" "$s"
    echo "$s  $name" >> "$WORK/SHA256SUMS"
  done
done
cp "$WORK/SHA256SUMS" "$DIST_DIR/SHA256SUMS"
SUMS_DIGEST="sha256:$(sha256 "$WORK/SHA256SUMS")"

# Plan: refuse before pushing anything if an exact-version tag holds other content.
: > "$WORK/conflicts"
plan_blob() { # <key> <tag> <want-layer-digest>
  local d have
  d="$(resolve "$2")"
  if [ -z "$d" ]; then
    kset "act|$1" push
    return
  fi
  kset "old|$1" "$d"
  have="$(layer_digest "$d")"
  if [ "$have" = "$3" ]; then
    kset "act|$1" skip
  elif [ "$REPLACE_EXISTING" = true ]; then
    kset "act|$1" replace
  else
    kset "act|$1" conflict
    echo "$2 holds $have, this build is $3" >> "$WORK/conflicts"
  fi
}
for c in "${CLOUD_LIST[@]}"; do
  for p in "${PLATFORM_LIST[@]}"; do
    plan_blob "$c|$p" "$VERSION-$c-${p/\//-}" "sha256:$(kget "sum|$c|$p")"
  done
done
plan_blob sums "$VERSION-checksums" "$SUMS_DIGEST"
for c in "${CLOUD_LIST[@]}"; do
  d="$(resolve "$VERSION-$c")"
  if [ -z "$d" ]; then
    kset "act|idx|$c" push
    continue
  fi
  kset "old|idx|$c" "$d"
  want=""
  for p in "${PLATFORM_LIST[@]}"; do
    if [ "$(kget "act|$c|$p")" = skip ]; then want+="$(kget "old|$c|$p")"$'\n'; else want+="new"$'\n'; fi
  done
  if [ "$(index_members "$d")" = "$(printf '%s' "$want" | sort)" ]; then
    kset "act|idx|$c" skip
  elif [ "$REPLACE_EXISTING" = true ]; then
    kset "act|idx|$c" replace
  else
    kset "act|idx|$c" conflict
    echo "$VERSION-$c is an index over other manifests" >> "$WORK/conflicts"
  fi
done
if [ -s "$WORK/conflicts" ]; then
  echo "refusing to publish $VERSION: exact-version tags already hold other content (set REPLACE_EXISTING=true to overwrite):" >&2
  sed 's/^/  /' "$WORK/conflicts" >&2
  exit 1
fi

if [ "$SIGN" = true ]; then
  : "${COSIGN_PRIVATE_KEY:?COSIGN_PRIVATE_KEY is required to sign (SIGN=true)}" "${COSIGN_PASSWORD:?COSIGN_PASSWORD is required to sign}"
  cosign public-key --key env://COSIGN_PRIVATE_KEY > "$WORK/cosign.pub"
  if [ -f "$PUBKEY_FILE" ] && ! cmp -s <(tr -d '\r' < "$PUBKEY_FILE") <(tr -d '\r' < "$WORK/cosign.pub"); then
    echo "COSIGN_PRIVATE_KEY does not match $PUBKEY_FILE" >&2
    exit 1
  fi
fi

annotations=(
  --annotation "org.opencontainers.image.version=$VERSION"
  --annotation "org.opencontainers.image.revision=${GIT_COMMIT:-unknown}"
  --annotation "org.opencontainers.image.source=${SOURCE_REPO:-unknown}"
  --annotation "org.opencontainers.image.created=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  --annotation "ai.rtifact.tofu-version=${TOFU_VERSION:-unknown}"
)
: > "$WORK/digests"
report() { # <action> <ref> <digest> [note]
  local a="$1"
  [ "$a" = push ] && a=pushed
  [ "$a" = replace ] && a=replaced
  [ "$a" = skip ] && a="already present"
  echo "$a: $2 $3${4:+ ($4)}"
}

# shellcheck disable=SC2086
for c in "${CLOUD_LIST[@]}"; do
  for p in "${PLATFORM_LIST[@]}"; do
    key="$c|$p"
    tag="$VERSION-$c-${p/\//-}"
    act="$(kget "act|$key")"
    if [ "$act" = skip ]; then
      d="$(kget "old|$key")"
    else
      stage="$WORK/stage-$c-${p/\//-}"
      mkdir -p "$stage"
      install -m 0755 "$(kget "file|$key")" "$stage/$PRODUCT"
      d="$(cd "$stage" && oras push $ORAS_FLAGS --artifact-type "$ARTIFACT_TYPE" --artifact-platform "$p" \
        "${annotations[@]}" --annotation "ai.rtifact.cloud=$c" --annotation "ai.rtifact.platform=$p" \
        --format go-template='{{.digest}}' "$REPO:$tag" "$PRODUCT:$ARTIFACT_TYPE")"
    fi
    kset "dig|$key" "$d"
    echo "$d" >> "$WORK/digests"
    report "$act" "$REPO:$tag" "$d"
  done
done

act="$(kget act\|sums)"
if [ "$act" = skip ]; then
  d="$(kget old\|sums)"
else
  # shellcheck disable=SC2086
  d="$(cd "$WORK" && oras push $ORAS_FLAGS --artifact-type "$SUMS_TYPE" "${annotations[@]}" \
    --format go-template='{{.digest}}' "$REPO:$VERSION-checksums" "SHA256SUMS:text/plain")"
fi
kset dig\|sums "$d"
echo "$d" >> "$WORK/digests"
report "$act" "$REPO:$VERSION-checksums" "$d"

for c in "${CLOUD_LIST[@]}"; do
  act="$(kget "act|idx|$c")"
  if [ "$act" = skip ]; then
    d="$(kget "old|idx|$c")"
  else
    members=()
    for p in "${PLATFORM_LIST[@]}"; do members+=("$(kget "dig|$c|$p")"); done
    # shellcheck disable=SC2086
    oras manifest index create $ORAS_FLAGS --artifact-type "$ARTIFACT_TYPE" \
      "${annotations[@]}" --annotation "ai.rtifact.cloud=$c" \
      "$REPO:$VERSION-$c" "${members[@]}" >/dev/null
    d="$(resolve "$VERSION-$c")"
  fi
  kset "dig|idx|$c" "$d"
  echo "$d" >> "$WORK/digests"
  report "$act" "$REPO:$VERSION-$c" "$d" index
done

if [ "$SIGN" = true ]; then
  # shellcheck disable=SC2086
  while read -r d; do
    ref="$REPO@$d"
    if cosign verify $COSIGN_REG --key "$WORK/cosign.pub" --insecure-ignore-tlog=true "$ref" >/dev/null 2>&1; then
      echo "already signed: $ref"
      continue
    fi
    if ! cosign sign --yes $COSIGN_REG --key env://COSIGN_PRIVATE_KEY --tlog-upload=false --use-signing-config=false "$ref" > "$WORK/sign.log" 2>&1; then
      echo "cosign sign failed for $ref:" >&2
      cat "$WORK/sign.log" >&2
      exit 1
    fi
    cosign verify $COSIGN_REG --key "$WORK/cosign.pub" --insecure-ignore-tlog=true "$ref" >/dev/null 2>&1 \
      || { echo "signature for $ref does not verify" >&2; exit 1; }
    echo "signed: $ref"
  done < <(sort -u "$WORK/digests")
fi

if [ "$PRERELEASE" = false ]; then
  for c in "${CLOUD_LIST[@]}"; do
    # shellcheck disable=SC2086
    oras tag $ORAS_FLAGS "$REPO@$(kget "dig|idx|$c")" "latest-$c" >/dev/null
    echo "latest-$c -> $VERSION-$c"
  done
fi

{
  echo "## Registry artifacts ($VERSION)"
  echo
  if [ "$REPLACE_EXISTING" = true ]; then
    echo "**replace_existing was set: exact-version tags may have been overwritten.**"
    echo
  fi
  echo "Repository: \`$REPO\`"
  echo
  echo "| Tag | Digest | Binary sha256 |"
  echo "|---|---|---|"
  for c in "${CLOUD_LIST[@]}"; do
    echo "| \`$VERSION-$c\` (index) | \`$(kget "dig|idx|$c")\` | |"
    for p in "${PLATFORM_LIST[@]}"; do
      echo "| \`$VERSION-$c-${p/\//-}\` | \`$(kget "dig|$c|$p")\` | \`$(kget "sum|$c|$p")\` |"
    done
  done
  echo "| \`$VERSION-checksums\` | \`$(kget dig\|sums)\` | |"
  if [ "$PRERELEASE" = false ]; then
    echo
    echo "\`latest-<cloud>\` points at the $VERSION index of each cloud."
  fi
  echo
  echo "Install (pick the cloud and platform):"
  echo
  echo '```bash'
  echo "oras login $REGISTRY -u <user>"
  echo "oras pull $REPO:$VERSION-${CLOUD_LIST[0]}-linux-amd64"
  echo "chmod +x $PRODUCT && ./$PRODUCT version"
  echo "cosign verify --key cosign.pub --insecure-ignore-tlog=true $REPO:$VERSION-${CLOUD_LIST[0]}-linux-amd64"
  echo '```'
} >> "$SUMMARY_FILE"
