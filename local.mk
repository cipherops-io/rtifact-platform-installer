# Installer product hooks — not managed by boilerplate.

.PHONY: validate-config bump-providers

VERSION_CLOUDS := azure aws gcp
VERSION_LAYERS := state-backend core-network data-platform app-platform network-edge ai-integrations

validate-config:
	@test -f $(DIST) || $(MAKE) build-dev
	@INSTALLER_CONTENT_ROOT=. ./$(DIST) validate examples/sandbox.yaml --check-azure=false

# After editing vars/versions.yml, sync pins into stack versions.tofu and refresh locks.
bump-providers:
	@command -v boilerplate >/dev/null || (echo "boilerplate CLI required" && exit 1)
	@set -e; \
	repo=$$(pwd); \
	profile=$$(grep '^ProfileID:' boilerplate-vars.yml | awk '{print $$2}'); \
	for c in $(VERSION_CLOUDS); do \
	  for l in $(VERSION_LAYERS); do \
	    target="$$c/$$l"; \
	    echo "==> versions $$target"; \
	    boilerplate \
	      --template-url "$$repo/build/boilerplate/stack-versions/layer-versions" \
	      --output-folder "$$repo/infra/stacks/$$profile/$$target" \
	      --non-interactive --disable-dependency-prompt --no-hooks \
	      --var-file boilerplate-vars.yml \
	      --var-file vars/versions.yml \
	      --var "StackTarget=$$target" \
	      --var "OpenTofuVersion=$(TOFU_VERSION)" \
	      --var 'ProviderSources={azure: "$(PROVIDER_SOURCE_AZURE)", aws: "$(PROVIDER_SOURCE_AWS)", gcp: "$(PROVIDER_SOURCE_GCP)"}'; \
	  done; \
	done
	@$(MAKE) tofu-lock

tofu-test-fast: package-content

tofu-test: validate-config