# DO NOT EDIT — managed by cipherops-io/boilerplates (rtifact-product).
# Change it there, then run: make boilerplate-update

# Generic boilerplate sync — rtifact-product managed template.

BOILERPLATES_TEMPLATE_URL ?= ../boilerplates/devops-templates/rtifact-product
# Before push / CI: github.com/cipherops-io/boilerplates//devops-templates/rtifact-product?ref=v3.0.0
# For copies outside the org workspace (e2e temp dir), override: make ... BOILERPLATES_VAR_REPO=/path/to/boilerplates
BOILERPLATES_VAR_REPO ?= ../boilerplates
BOILERPLATE_VAR_FILES ?= boilerplate-vars.yml $(BOILERPLATES_VAR_REPO)/vars/versions.yml $(BOILERPLATES_VAR_REPO)/vars/layers.yml

ifndef REPO_ROOT
REPO_ROOT := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
endif

.PHONY: boilerplate-update boilerplate-verify check-template-source print-tool-versions

boilerplate-update:
	@command -v boilerplate >/dev/null || (echo "boilerplate CLI required (https://boilerplate.gruntwork.io)" && exit 1)
	@set -e; \
	args=""; \
	for f in $(BOILERPLATE_VAR_FILES); do \
	  [ -f "$$f" ] || { echo "missing var file: $$f"; exit 1; }; \
	  args="$$args --var-file $$f"; \
	done; \
	boilerplate \
	  --template-url $(BOILERPLATES_TEMPLATE_URL) \
	  --non-interactive --disable-dependency-prompt --no-hooks \
	  --output-folder . \
	  $$args; \
	go mod tidy

boilerplate-verify:
	@set -e; \
	tmp=$$(mktemp -d); \
	trap 'rm -rf "$$tmp"' EXIT; \
	$(MAKE) -f $(firstword $(MAKEFILE_LIST)) boilerplate-update BOILERPLATES_TEMPLATE_URL=$(BOILERPLATES_TEMPLATE_URL) REPO_ROOT="$$tmp" 2>/dev/null || true; \
	test -f .rtifact/managed-files.txt; \
	while IFS= read -r path; do \
	  [ -z "$$path" ] && continue; \
	  case "$$path" in \#*) continue;; esac; \
	  [ -f "$$path" ] || continue; \
	  if [ -f "$$tmp/$$path" ]; then \
	    diff -q "$$path" "$$tmp/$$path" >/dev/null || { echo "managed drift: $$path"; diff -u "$$path" "$$tmp/$$path" | head -20; exit 1; }; \
	  fi; \
	done < .rtifact/managed-files.txt; \
	echo "boilerplate-verify: OK"

check-template-source:
	@case "$(BOILERPLATES_TEMPLATE_URL)" in \
	  ../*|/*|~/*|file:*) echo "local template path forbidden in CI — use github.com/...?ref=v3.0.0"; exit 1;; \
	esac
	@echo "check-template-source: OK"

print-tool-versions:
	@echo "TOFU_VERSION=$(TOFU_VERSION)"
	@echo "TERRAFORM_DOCS_VERSION=$(TERRAFORM_DOCS_VERSION)"
	@echo "BOILERPLATE_VERSION=$(BOILERPLATE_VERSION)"
# DO NOT EDIT — managed by cipherops-io/boilerplates (rtifact-product).
# Change it there, then run: make boilerplate-update

# Scaffold targets — render skeleton templates once (init-only skip rules).

BOILERPLATES_GITHUB ?= github.com/cipherops-io/boilerplates
BOILERPLATES_REF ?= ?ref=v3.0.0
BOILERPLATES_REPO ?= ../boilerplates
TPL_RTIFACT_STACK ?= $(BOILERPLATES_GITHUB)//devops-templates/rtifact-stack$(BOILERPLATES_REF)
TPL_RTIFACT_MODULE ?= $(BOILERPLATES_GITHUB)//devops-templates/rtifact-module$(BOILERPLATES_REF)

.PHONY: scaffold-module scaffold-stack-layer

.check-boilerplate-cli:
	@command -v boilerplate >/dev/null || (echo "boilerplate CLI required" && exit 1)

scaffold-module: .check-boilerplate-cli
	@test -n "$(CLOUD)" && test -n "$(MODULE)" || { echo "usage: make scaffold-module CLOUD=azure MODULE=my-capability"; exit 1; }
	@set -e; \
	tpl="$(TPL_RTIFACT_MODULE)"; \
	[ -d "$(BOILERPLATES_REPO)/devops-templates/rtifact-module" ] && tpl="$(BOILERPLATES_REPO)/devops-templates/rtifact-module"; \
	args=""; \
	for f in $(BOILERPLATE_VAR_FILES); do args="$$args --var-file $$f"; done; \
	boilerplate --template-url "$$tpl" \
	  --output-folder "$(REPO_ROOT)/infra/modules/$(CLOUD)/$(MODULE)" \
	  --non-interactive --disable-dependency-prompt --no-hooks \
	  $$args --var Cloud=$(CLOUD) --var ModuleName=$(MODULE)

scaffold-stack-layer: .check-boilerplate-cli
	@test -n "$(CLOUD)" && test -n "$(LAYER)" || { echo "usage: make scaffold-stack-layer CLOUD=azure LAYER=my-layer [PROFILE=<profile-id>]"; exit 1; }
	@set -e; \
	tpl="$(TPL_RTIFACT_STACK)"; \
	[ -d "$(BOILERPLATES_REPO)/devops-templates/rtifact-stack" ] && tpl="$(BOILERPLATES_REPO)/devops-templates/rtifact-stack"; \
	profile="$${PROFILE:-$(PROFILES)}"; \
	args=""; \
	for f in $(BOILERPLATE_VAR_FILES); do args="$$args --var-file $$f"; done; \
	boilerplate --template-url "$$tpl" \
	  --output-folder "$(REPO_ROOT)/infra/stacks/$$profile/$(CLOUD)/$(LAYER)" \
	  --non-interactive --disable-dependency-prompt --no-hooks \
	  $$args --var Cloud=$(CLOUD) --var Layer=$(LAYER) --var ProfileID=$$profile
# DO NOT EDIT — managed by cipherops-io/boilerplates (rtifact-product).
# Change it there, then run: make boilerplate-update

include $(dir $(lastword $(MAKEFILE_LIST)))tooling.mk

TERRAFORM_DOCS_CONFIG ?= $(REPO_ROOT)/.terraform-docs.yml
TERRAFORM_DOCS_GEN_CONFIG ?= $(REPO_ROOT)/.terraform-docs.gen.yml

CLOUDS ?= azure,aws,gcp
PROFILES ?= aios-infra

.PHONY: help check-tools
.PHONY: tofu-fmt tofu-fmt-check tofu-docs tofu-docs-check
.PHONY: tofu-validate tofu-lock tofu-providers-lock
.PHONY: tofu-test-fast tofu-test tofu-pre-commit tofu-pre-commit-fast
.PHONY: go-fmt go-vet go-test go-tidy

help:
	@grep -E '^[a-zA-Z0-9_.-]+:' $(MAKEFILE_LIST) | cut -d: -f1 | sort -u

check-tools:
	@command -v tofu >/dev/null || { echo "missing: tofu"; exit 1; }
	@command -v go >/dev/null || { echo "missing: go"; exit 1; }
	@command -v boilerplate >/dev/null || { echo "missing: boilerplate"; exit 1; }
	@command -v terraform-docs >/dev/null || echo "warn: terraform-docs not on PATH"
	@echo "check-tools: OK"

tofu-fmt:
	@tofu fmt -recursive infra/

tofu-fmt-check:
	@tofu fmt -check -diff -recursive infra/

tofu-docs:
	@set -e; \
	for d in $$(find "$(REPO_ROOT)/infra/modules" -mindepth 2 -maxdepth 2 -type f -name providers.tofu -exec dirname {} \;); do \
	  echo "==> docs $$d"; \
	  (cd "$$d" && "$(REPO_ROOT)/build/scripts/terraform-docs-module.sh" "$(TERRAFORM_DOCS_CONFIG)"); \
	done

tofu-docs-check:
	@set -e; \
	for d in $$(find "$(REPO_ROOT)/infra/modules" -mindepth 2 -maxdepth 2 -type f -name providers.tofu -exec dirname {} \;); do \
	  (cd "$$d" && "$(REPO_ROOT)/build/scripts/terraform-docs-module.sh" "$(TERRAFORM_DOCS_CONFIG)" --output-check); \
	done

tofu-validate:
	@set -e; IFS=','; \
	for c in $(CLOUDS); do \
	  c=$$(printf '%s' "$$c" | tr -d ' '); \
	  for d in $$(find "$(REPO_ROOT)/infra/stacks" -path "*/$$c/*" \( -name providers.tofu -o -name versions.tofu \) -exec dirname {} \; | sort -u); do \
	    echo "==> $$d"; \
	    (cd "$$d" && tofu init -backend=false -input=false && tofu validate); \
	  done; \
	done

tofu-lock tofu-providers-lock:
	@set -e; IFS=','; \
	for p in $(LOCK_PLATFORMS); do \
	  for c in $(CLOUDS); do \
	    c=$$(printf '%s' "$$c" | tr -d ' '); \
	    for d in $$(find "$(REPO_ROOT)/infra/stacks" -path "*/$$c/*" \( -name providers.tofu -o -name versions.tofu \) -exec dirname {} \; | sort -u); do \
	      echo "==> lock $$d ($$p)"; \
	      (cd "$$d" && tofu providers lock -platform=$$p); \
	    done; \
	  done; \
	done

tofu-test-fast: tofu-fmt-check check-tools
	@set -e; \
	for d in $$(find "$(REPO_ROOT)/infra/modules" -mindepth 2 -maxdepth 2 -type f -name providers.tofu -exec dirname {} \;); do \
	  echo "==> module $$d"; \
	  (cd "$$d" && tofu fmt -check -diff && tofu init -backend=false -input=false && tofu validate); \
	done; \
	$(MAKE) go-test

tofu-test: tofu-test-fast tofu-validate
	@$(MAKE) -f $(firstword $(MAKEFILE_LIST)) validate-config 2>/dev/null || true

tofu-pre-commit-fast: tofu-fmt tofu-docs go-fmt go-tidy

tofu-pre-commit: tofu-pre-commit-fast tofu-lock

go-fmt:
	@go fmt ./...

go-vet:
	@go vet ./...

go-test:
	@go test ./...

go-tidy:
	@go mod tidy

-include local.mk