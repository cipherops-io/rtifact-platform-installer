# DO NOT EDIT — managed from boilerplates/partials/makefile/ (render: cd boilerplates && make sync-installer-managed)

# Managed makefile fragment — merged into generated boilerplate.mk (do not edit product copy by hand).

BOILERPLATES_TEMPLATE_URL ?= github.com/cipherops-io/boilerplates//devops-templates/platform-installer
# Local sibling: make boilerplate-update BOILERPLATES_TEMPLATE_URL=../boilerplates/devops-templates/platform-installer
BOILERPLATE_VAR_FILES ?= boilerplate-vars.yml
ifndef REPO_ROOT
REPO_ROOT := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
endif

.PHONY: boilerplate-update
.PHONY: .check-tofu-only .check-tofu-tools .check-tofu-tools-fast

.check-tofu-only:
	@command -v tofu >/dev/null || (echo "tofu required" && exit 1)

.check-tofu-tools:
	@command -v tofu >/dev/null || (echo "tofu required; install OpenTofu $(TOFU_VERSION)" && exit 1)
	@command -v terraform-docs >/dev/null || (echo "terraform-docs required; see docs/BOILERPLATE.md or BOILERPLATE.md" && exit 1)

.check-tofu-tools-fast:
	@command -v tofu >/dev/null || (echo "tofu required" && exit 1)
	@command -v terraform-docs >/dev/null || (echo "terraform-docs required" && exit 1)

boilerplate-update:
	@command -v boilerplate >/dev/null || (echo "boilerplate CLI required (https://boilerplate.gruntwork.io)" && exit 1)
	@set -e; \
	files="$(BOILERPLATE_VAR_FILES)"; \
	args=""; \
	for f in $$files; do \
	  [ -f "$$f" ] || { echo "missing var file: $$f"; exit 1; }; \
	  args="$$args --var-file $$f"; \
	done; \
	boilerplate \
	  --template-url $(BOILERPLATES_TEMPLATE_URL) \
	  --non-interactive \
	  --disable-dependency-prompt \
	  --no-hooks \
	  --output-folder . \
	  $$args; \
	go mod tidy
# Managed makefile fragment — IaC + Go targets for platform-installer (multi-cloud).

PACKAGED_CLOUDS ?= azure,aws,gcp
BINARY ?= rtifact
DIST ?= dist/$(BINARY)
TERRAFORM_DOCS_CONFIG ?= $(REPO_ROOT)/.terraform-docs.yml
TERRAFORM_DOCS_GEN_CONFIG ?= $(REPO_ROOT)/.terraform-docs.gen.yml

export REPO_ROOT

.PHONY: tofu-fmt do/tofu-fmt tofu-docs do/tofu-docs
.PHONY: tofu-validate tofu-test tofu-test-fast do/tofu-test do/tofu-test-fast
.PHONY: tofu-pre-commit tofu-pre-commit-fast
.PHONY: tofu-providers-lock do/tofu-providers-lock

do/tofu-fmt: .check-tofu-tools-fast
	tofu fmt

tofu-fmt: .check-tofu-tools-fast
	tofu fmt -recursive infra/

do/tofu-docs: .check-tofu-tools
	@test -f "$(TERRAFORM_DOCS_CONFIG)" || { echo "missing $(TERRAFORM_DOCS_CONFIG)"; exit 1; }
	@"$(REPO_ROOT)/build/scripts/terraform-docs-module.sh" "$(TERRAFORM_DOCS_CONFIG)"

tofu-docs: .check-tofu-tools
	@set -e; \
	for d in infra/modules/*/*/; do \
	  [ -f "$$d/providers.tofu" ] || continue; \
	  echo "==> docs $$d"; \
	  (cd "$$d" && "$(REPO_ROOT)/build/scripts/terraform-docs-module.sh" "$(TERRAFORM_DOCS_CONFIG)"); \
	done

tofu-validate: .check-tofu-only
	@set -e; CLOUDS="$(PACKAGED_CLOUDS)"; IFS=','; \
	for c in $$CLOUDS; do \
	  c=$$(printf '%s' "$$c" | tr -d ' '); \
	  for d in infra/stacks/aios-infra/$$c/*/; do \
	    [ -f "$$d/providers.tofu" ] || continue; \
	    echo "==> $$d"; (cd "$$d" && tofu init -backend=false -input=false && tofu validate); \
	  done; \
	done

do/tofu-test-fast: .check-tofu-tools-fast
	@test -f providers.tofu || exit 0
	@set -e; tofu fmt -check -diff; tofu init -backend=false -input=false; tofu validate

tofu-test-fast: .check-tofu-tools-fast
	@set -e; \
	for readme in infra/modules/*/*/README.md; do \
	  [ -f "$$readme" ] || continue; \
	  grep -q 'No inputs' "$$readme" && { echo "stale terraform-docs in $$readme — run: make tofu-docs" >&2; exit 1; }; \
	done; \
	tofu fmt -check -recursive infra/; \
	for d in infra/modules/*/*/; do \
	  [ -f "$$d/providers.tofu" ] || continue; \
	  echo "==> test-fast $$d"; \
	  (cd "$$d" && set -e && tofu fmt -check -diff && tofu init -backend=false -input=false && tofu validate); \
	  if [ -f "$$d/README.md" ] && [ -f "$(TERRAFORM_DOCS_CONFIG)" ]; then \
	    (cd "$$d" && "$(REPO_ROOT)/build/scripts/terraform-docs-module.sh" "$(TERRAFORM_DOCS_CONFIG)" --output-check); \
	  fi; \
	done; \
	chmod +x "$(REPO_ROOT)/build/scripts/"*.sh; \
	$(MAKE) package-content; \
	go test ./...

do/tofu-test: .check-tofu-tools
	@test -f providers.tofu || exit 0
	@$(MAKE) do/tofu-test-fast
	@if [ -f .terraform.lock.hcl ]; then \
	  tofu init -backend=false -input=false -lockfile=readonly; \
	fi

tofu-test: .check-tofu-tools
	@$(MAKE) tofu-test-fast
	@$(MAKE) tofu-validate
	@test -f "$(DIST)" || $(MAKE) build-dev
	@INSTALLER_CONTENT_ROOT=. ./$(DIST) validate examples/sandbox.yaml --check-azure=false

do/tofu-providers-lock: .check-tofu-tools
	@test -f providers.tofu || exit 0
	tofu providers lock \
		-platform=linux_amd64 \
		-platform=linux_arm64 \
		-platform=darwin_amd64 \
		-platform=darwin_arm64 \
		-platform=windows_amd64

tofu-providers-lock: .check-tofu-tools
	@set -e; CLOUDS="$(PACKAGED_CLOUDS)"; IFS=','; \
	for c in $$CLOUDS; do \
	  c=$$(printf '%s' "$$c" | tr -d ' '); \
	  for d in infra/stacks/aios-infra/$$c/*/; do \
	    [ -f "$$d/providers.tofu" ] || continue; \
	    echo "==> lock $$d"; \
	    (cd "$$d" && tofu init -backend=false -input=false && tofu providers lock \
	      -platform=linux_amd64 -platform=linux_arm64 \
	      -platform=darwin_amd64 -platform=darwin_arm64 \
	      -platform=windows_amd64); \
	  done; \
	done

tofu-pre-commit-fast: .check-tofu-tools
	@$(MAKE) tofu-fmt tofu-docs
	@$(MAKE) fmt
	@go mod tidy

tofu-pre-commit: .check-tofu-tools
	@$(MAKE) tofu-pre-commit-fast
	@$(MAKE) tofu-providers-lock