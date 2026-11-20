#!/usr/bin/env bash
# Permanent OpenTofu module docs: terraform-docs still inspects .tf only (not .tofu metadata).
# We render from a temp copy of *.tofu → *.tf, then splice into README inject markers — no .tf in the repo.
set -euo pipefail

module_root=$(pwd)
config="${1:?repo-root .terraform-docs.yml (inject config)}"
output_check=false
if [ "${2:-}" = "--output-check" ]; then
  output_check=true
fi

repo_root="${REPO_ROOT:-$(cd "$(dirname "$0")/../.." && pwd)}"
gen_config="${TERRAFORM_DOCS_GEN_CONFIG:-$repo_root/.terraform-docs.gen.yml}"
readme="$module_root/README.md"

[ -f "$config" ] || { echo "missing config: $config" >&2; exit 1; }
[ -f "$gen_config" ] || { echo "missing gen config: $gen_config" >&2; exit 1; }
[ -f "$readme" ] || { echo "missing README.md in $module_root" >&2; exit 1; }
grep -q 'BEGIN_TF_DOCS' "$readme" || { echo "README.md missing BEGIN_TF_DOCS markers: $readme" >&2; exit 1; }

tmpdir=$(mktemp -d)
cleanup() { rm -rf "$tmpdir"; }
trap cleanup EXIT ERR INT

shopt -s nullglob
tofu_files=(*.tofu)
[ ${#tofu_files[@]} -gt 0 ] || { echo "no *.tofu in $module_root" >&2; exit 1; }

for f in "${tofu_files[@]}"; do
  cp "$f" "$tmpdir/${f%.tofu}.tf"
done

terraform-docs markdown "$tmpdir" --config "$gen_config" >/dev/null
fragment="$tmpdir/_module_docs_fragment.md"
[ -s "$fragment" ] || { echo "terraform-docs produced empty fragment for $module_root" >&2; exit 1; }

if grep -q '^## Inputs$' "$fragment" && grep -A2 '^## Inputs$' "$fragment" | grep -q 'No inputs'; then
  echo "terraform-docs returned empty inputs for $module_root (check .tofu / temp generation)" >&2
  exit 1
fi

new_readme=$(mktemp)
trap 'rm -rf "$tmpdir" "$new_readme"' EXIT ERR INT

awk -v frag="$fragment" '
  BEGIN {
    while ((getline line < frag) > 0) {
      if (line ~ /^<!-- (BEGIN|END)_TF_DOCS -->$/) {
        continue
      }
      body = body line "\n"
    }
  }
  /^<!-- BEGIN_TF_DOCS -->$/ {
    if (!injected) {
      print
      printf "%s", body
      skip = 1
      injected = 1
    }
    next
  }
  /^<!-- END_TF_DOCS -->$/ {
    if (skip) {
      skip = 0
      print
    }
    next
  }
  skip { next }
  { print }
' "$readme" > "$new_readme"

if $output_check; then
  if ! cmp -s "$readme" "$new_readme"; then
    echo "README.md is out of date: $readme" >&2
    exit 1
  fi
else
  mv "$new_readme" "$readme"
  echo "README.md updated successfully"
fi