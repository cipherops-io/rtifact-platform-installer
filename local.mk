# Installer product hooks — not managed by boilerplate.

.PHONY: validate-config bump-providers

validate-config:
	@test -f $(DIST) || $(MAKE) build-dev
	@INSTALLER_CONTENT_ROOT=. ./$(DIST) validate examples/sandbox.yaml --check-azure=false

# After editing vars/versions.yml, sync pins into stack versions.tofu and refresh locks.
bump-providers:
	@python3 <<-'PY'
	import pathlib, re, yaml
	root = pathlib.Path(".")
	data = yaml.safe_load((root / "vars/versions.yml").read_text())
	pins = {**data.get("ProviderVersions", {}), **data.get("ExtraProviderVersions", {})}
	short = {"azurerm": pins.get("azure"), "aws": pins.get("aws"), "google": pins.get("gcp")}
	pins = {**pins, **{k: v for k, v in short.items() if v}}
	for vf in sorted(root.glob("infra/stacks/**/versions.tofu")):
	    text = vf.read_text()
	    for name, ver in pins.items():
	        if not ver:
	            continue
	        text, _ = re.subn(
	            rf"({re.escape(name)}\s*=\s*{{[^}}]*version\s*=\s*)\"[^\"]+\"",
	            rf'\1"{ver}"',
	            text,
	        )
	    vf.write_text(text)
	    print("bump", vf)
	PY
	@$(MAKE) tofu-lock

tofu-test-fast: package-content

tofu-test: validate-config