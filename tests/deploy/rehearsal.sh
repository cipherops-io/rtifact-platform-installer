#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
BIN="${RTIFACT_BIN:-$ROOT/dist/rtifact}"
if [ ! -x "$BIN" ]; then
  echo "rehearsal: dist/rtifact missing (run: make build-dev)" >&2
  exit 1
fi
export INSTALLER_CONTENT_ROOT="$ROOT"
"$BIN" version
"$BIN" deploy validate "$ROOT/examples/deploy/minimal.yaml"
"$BIN" deploy validate "$ROOT/examples/deploy/demo/demo.yaml"
"$BIN" deploy plan --help | grep -q -- '--adopt'
"$BIN" deploy status --help | grep -q -- '\-o'
"$BIN" deploy plan --help | grep -q -- 'json'
"$BIN" deploy apply --help | grep -q -- '--yes'
"$BIN" deploy destroy --help | grep -q -- 'deployment'
echo "rehearsal OK"