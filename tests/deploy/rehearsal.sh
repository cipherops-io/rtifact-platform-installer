#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
BIN="${RTIFACT_BIN:-$ROOT/dist/rtifact}"
if [ ! -x "$BIN" ]; then
  echo "rehearsal: skip (build dist/rtifact first)"
  exit 0
fi
export INSTALLER_CONTENT_ROOT="$ROOT"
"$BIN" deploy validate "$ROOT/examples/deploy/minimal.yaml"
"$BIN" deploy validate "$ROOT/examples/deploy/demo/demo.yaml"
echo "rehearsal OK"