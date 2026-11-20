#!/usr/bin/env bash
# Rewrite provider version pins in stack versions.tofu from vars/versions.yml
set -euo pipefail
root="$(cd "$(dirname "$0")/.." && pwd)"
vars="$root/vars/versions.yml"
[ -f "$vars" ] || { echo "missing $vars"; exit 1; }

python3 - "$root" "$vars" <<'PY'
import re, sys, pathlib, yaml
root, vars = pathlib.Path(sys.argv[1]), pathlib.Path(sys.argv[2])
data = yaml.safe_load(vars.read_text())
pv = data.get("ProviderVersions", {})
extra = data.get("ExtraProviderVersions", {})
for vf in root.glob("infra/stacks/**/versions.tofu"):
    text = vf.read_text()
    for name, ver in {**pv, **extra}.items():
        pat = rf'({name}\s*=\s*{{[^}}]*version\s*=\s*)"[^"]+"'
        text, n = re.subn(pat, rf'\1"{ver}"', text)
    vf.write_text(text)
    print("bump", vf)
PY