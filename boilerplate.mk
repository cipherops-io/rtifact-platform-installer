# DO NOT EDIT — managed by cipherops-io/boilerplates (rtifact-product).
# Change it there, then run: make boilerplate-update

# Generic boilerplate sync — rtifact-product managed template.

BOILERPLATES_TEMPLATE_URL ?= github.com/cipherops-io/boilerplates//devops-templates/rtifact-product?ref=v0.0.1
# Local sibling: make boilerplate-update BOILERPLATES_TEMPLATE_URL=../boilerplates/devops-templates/rtifact-product
BOILERPLATES_REPO ?= ../boilerplates
BOILERPLATES_REF ?= v0.0.1
BOILERPLATE_VAR_FILES ?= boilerplate-vars.yml

ifndef REPO_ROOT
REPO_ROOT := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
endif

.PHONY: boilerplate-update boilerplate-verify check-template-source print-tool-versions
.PHONY: record-boilerplate-source verify-boilerplate-source

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
	  --manifest-file boilerplate-manifest.yaml \
	  --output-folder . \
	  $$args; \
	$(MAKE) -f $(firstword $(MAKEFILE_LIST)) record-boilerplate-source

record-boilerplate-source:
	@set -e; \
	ref="$(BOILERPLATES_REF)"; \
	url="$(BOILERPLATES_TEMPLATE_URL)"; \
	case "$$url" in \
	  ../*|file:*|$(BOILERPLATES_REPO)/*|$(BOILERPLATES_REPO)) \
	    commit=$$(git -C "$(BOILERPLATES_REPO)" rev-parse HEAD); \
	    ;; \
	  *) \
	    if [ -d "$(BOILERPLATES_REPO)/.git" ]; then \
	      commit=$$(git -C "$(BOILERPLATES_REPO)" rev-parse "refs/tags/$$ref^{commit}" 2>/dev/null \
	        || git -C "$(BOILERPLATES_REPO)" rev-parse "$$ref^{commit}" 2>/dev/null); \
	    fi; \
	    if [ -z "$$commit" ]; then \
	      remote="https://github.com/cipherops-io/boilerplates.git"; \
	      commit=$$(git ls-remote "$$remote" "refs/tags/$$ref^{}" 2>/dev/null | awk '{print $$1}'); \
	    fi; \
	    [ -n "$$commit" ] || { echo "record-boilerplate-source: cannot resolve ref $$ref"; exit 1; }; \
	    ;; \
	esac; \
	printf '%s\n' \
	  "# DO NOT EDIT — managed by cipherops-io/boilerplates (rtifact-product)." \
	  "# commit is refreshed by make boilerplate-update (record-boilerplate-source)." \
	  "ref: $$ref" \
	  "commit: $$commit" \
	  "template_url: $$url" > boilerplate-source.yaml

verify-boilerplate-source:
	@set -e; \
	case "$(BOILERPLATES_TEMPLATE_URL)" in \
	  ../*|file:*|$(BOILERPLATES_REPO)/*|$(BOILERPLATES_REPO)) \
	    echo "verify-boilerplate-source: skip (local template)"; exit 0;; \
	esac; \
	[ -f boilerplate-source.yaml ] || { echo "missing boilerplate-source.yaml"; exit 1; }; \
	ref=$$(awk '/^ref:/{print $$2; exit}' boilerplate-source.yaml); \
	want=$$(awk '/^commit:/{print $$2; exit}' boilerplate-source.yaml); \
	[ -n "$$ref" ] && [ -n "$$want" ] || { echo "boilerplate-source.yaml missing ref or commit"; exit 1; }; \
	got=""; \
	if [ -d "$(BOILERPLATES_REPO)/.git" ]; then \
	  got=$$(git -C "$(BOILERPLATES_REPO)" rev-parse "refs/tags/$$ref^{commit}" 2>/dev/null \
	    || git -C "$(BOILERPLATES_REPO)" rev-parse "$$ref^{commit}" 2>/dev/null); \
	fi; \
	if [ -z "$$got" ]; then \
	  remote="https://github.com/cipherops-io/boilerplates.git"; \
	  got=$$(git ls-remote "$$remote" "refs/tags/$$ref^{}" 2>/dev/null | awk '{print $$1}'); \
	fi; \
	[ -n "$$got" ] || { echo "verify-boilerplate-source: cannot resolve tag $$ref"; exit 1; }; \
	if [ "$$got" != "$$want" ]; then \
	  echo "boilerplate tag $$ref moved: recorded $$want, tag is now $$got"; \
	  echo "Run make boilerplate-update with the intended ref and commit boilerplate-source.yaml."; \
	  exit 1; \
	fi; \
	echo "verify-boilerplate-source: OK"

boilerplate-verify:
	@set -e; \
	$(MAKE) -f $(firstword $(MAKEFILE_LIST)) verify-boilerplate-source; \
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
	@echo "KIND_VERSION=$(KIND_VERSION)"
# DO NOT EDIT — managed by cipherops-io/boilerplates (rtifact-product).
# Change it there, then run: make boilerplate-update

# Scaffold targets — render skeleton templates once (init-only skip rules).

BOILERPLATES_GITHUB ?= github.com/cipherops-io/boilerplates
BOILERPLATES_GITHUB_REF ?= ?ref=v0.0.1
BOILERPLATES_REPO ?= ../boilerplates
TPL_RTIFACT_STACK ?= $(BOILERPLATES_GITHUB)//devops-templates/rtifact-stack$(BOILERPLATES_GITHUB_REF)
TPL_RTIFACT_MODULE ?= $(BOILERPLATES_GITHUB)//devops-templates/rtifact-module$(BOILERPLATES_GITHUB_REF)
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
# Profile ids discovered from profiles/*/profile.yaml (override: make PROFILES=id1 id2).
PROFILES ?= $(shell set -e; for f in "$(REPO_ROOT)"/profiles/*/profile.yaml; do \
	[ -f "$$f" ] || continue; basename "$$(dirname "$$f")"; \
	done | tr '\n' ' ' | sed 's/[[:space:]]*$$//')

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