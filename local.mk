# Installer product hooks — not managed by boilerplate.

.PHONY: validate-config bump-providers

validate-config:
	@test -f $(DIST) || $(MAKE) build-dev
	@INSTALLER_CONTENT_ROOT=. ./$(DIST) validate examples/sandbox.yaml --check-azure=false

bump-providers:
	@chmod +x scripts/bump-providers.sh
	@scripts/bump-providers.sh
	@$(MAKE) tofu-lock

tofu-test-fast: package-content

tofu-test: validate-config