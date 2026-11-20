# DO NOT EDIT — managed by cipherops-io/boilerplates (rtifact-product).
# Change it there, then run: make boilerplate-update

# Generic boilerplate sync — rtifact-product managed template.

BOILERPLATES_TEMPLATE_URL ?= github.com/cipherops-io/boilerplates//devops-templates/rtifact-product?ref=v3.0.0
# Local sibling: make boilerplate-update BOILERPLATES_TEMPLATE_URL=../boilerplates/devops-templates/rtifact-product
BOILERPLATES_REPO ?= ../boilerplates
BOILERPLATES_REF ?= v3.0.0
BOILERPLATE_VAR_FILES ?= boilerplate-vars.yml
_BUILTIN_DEFAULTS := $(REPO_ROOT)/build/boilerplate/defaults
_TPL_DEFAULTS_LOCAL := $(BOILERPLATES_REPO)/devops-templates/rtifact-product/defaults

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
	if [ -f "$(_BUILTIN_DEFAULTS)/layers.yml" ] && [ -f "$(_BUILTIN_DEFAULTS)/tool-versions.yml" ]; then \
	  args="$$args --var-file $(_BUILTIN_DEFAULTS)/layers.yml --var-file $(_BUILTIN_DEFAULTS)/tool-versions.yml"; \
	elif [ -f "$(_TPL_DEFAULTS_LOCAL)/layers.yml" ] && [ -f "$(_TPL_DEFAULTS_LOCAL)/tool-versions.yml" ]; then \
	  args="$$args --var-file $(_TPL_DEFAULTS_LOCAL)/layers.yml --var-file $(_TPL_DEFAULTS_LOCAL)/tool-versions.yml"; \
	fi; \
	boilerplate \
	  --template-url $(BOILERPLATES_TEMPLATE_URL) \
	  --non-interactive --disable-dependency-prompt --no-hooks \
	  --manifest-file boilerplate-manifest.yaml \
	  --output-folder . \
	  $$args

boilerplate-verify:
	@set -e; \
	$(MAKE) -f $(firstword $(MAKEFILE_LIST)) boilerplate-update; \
	git diff --exit-code -- . ':(exclude)boilerplate-manifest.yaml'

check-template-source:
	@case "$(BOILERPLATES_TEMPLATE_URL)" in \
	  ../*|/*|~/*|file:*) echo "local template path forbidden in CI — use github.com/...?ref=$(BOILERPLATES_REF)"; exit 1;; \
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
BOILERPLATES_GITHUB_REF ?= ?ref=v3.0.0
BOILERPLATES_REPO ?= ../boilerplates
TPL_RTIFACT_STACK ?= $(BOILERPLATES_GITHUB)//devops-templates/rtifact-stack$(BOILERPLATES_GITHUB_REF)
TPL_RTIFACT_MODULE ?= $(BOILERPLATES_GITHUB)//devops-templates/rtifact-module$(BOILERPLATES_GITHUB_REF)
_TPL_DEFAULTS_LOCAL := $(BOILERPLATES_REPO)/devops-templates/rtifact-product/defaults
_SCAFFOLD_BAKED_VARS := --var OpenTofuVersion=1.13.9

.PHONY: scaffold-module scaffold-stack-layer

.check-boilerplate-cli:
	@command -v boilerplate >/dev/null || (echo "boilerplate CLI required" && exit 1)

scaffold-module: .check-boilerplate-cli
	@test -n "$(CLOUD)" && test -n "$(MODULE)" || { echo "usage: make scaffold-module CLOUD=azure MODULE=my-capability"; exit 1; }
	@set -e; \
	tpl="$(TPL_RTIFACT_MODULE)"; \
	[ -d "$(BOILERPLATES_REPO)/devops-templates/rtifact-module" ] && tpl="$(BOILERPLATES_REPO)/devops-templates/rtifact-module"; \
	args="--var-file boilerplate-vars.yml $(_SCAFFOLD_BAKED_VARS)"; \
	[ -f "$(_TPL_DEFAULTS_LOCAL)/layers.yml" ] && args="$$args --var-file $(_TPL_DEFAULTS_LOCAL)/layers.yml"; \
	[ -f "$(_TPL_DEFAULTS_LOCAL)/tool-versions.yml" ] && args="$$args --var-file $(_TPL_DEFAULTS_LOCAL)/tool-versions.yml"; \
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
	args="--var-file boilerplate-vars.yml $(_SCAFFOLD_BAKED_VARS)"; \
	[ -f "$(_TPL_DEFAULTS_LOCAL)/layers.yml" ] && args="$$args --var-file $(_TPL_DEFAULTS_LOCAL)/layers.yml"; \
	[ -f "$(_TPL_DEFAULTS_LOCAL)/tool-versions.yml" ] && args="$$args --var-file $(_TPL_DEFAULTS_LOCAL)/tool-versions.yml"; \
	boilerplate --template-url "$$tpl" \
	  --output-folder "$(REPO_ROOT)/infra/stacks/$$profile/$(CLOUD)/$(LAYER)" \
	  --non-interactive --disable-dependency-prompt --no-hooks \
	  $$args --var Cloud=$(CLOUD) --var Layer=$(LAYER) --var ProfileID=$$profile
# DO NOT EDIT — managed by cipherops-io/boilerplates (rtifact-product).
# Change it there, then run: make boilerplate-update

include $(dir $(lastword $(MAKEFILE_LIST)))tooling.mk

TERRAFORM_DOCS_CONFIG ?= $(REPO_ROOT)/.terraform-docs.yml
TERRAFORM_DOCS_GEN_CONFIG ?= $(REPO_ROOT)/.terraform-docs.gen.yml

CLOUDS ?= azure aws gcp
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
	@set -e; \
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