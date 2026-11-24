MODULE := github.com/cipherops-io/rtifact-platform-installer
BINARY := rtifact
DIST := dist/$(BINARY)
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
PACKAGED_CLOUDS ?= azure,aws,gcp
VERSION ?= 0.0.1-dev
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
LDFLAGS = -s -w \
  -X github.com/cipherops-io/rtifact-platform-installer/internal/version.Release=true \
  -X github.com/cipherops-io/rtifact-platform-installer/internal/version.PackagedClouds=$(PACKAGED_CLOUDS) \
  -X github.com/cipherops-io/rtifact-platform-installer/internal/version.Version=$(VERSION) \
  -X github.com/cipherops-io/rtifact-platform-installer/internal/version.GitCommit=$(GIT_COMMIT) \
  -X github.com/cipherops-io/rtifact-platform-installer/internal/version.TofuVersion=$(TOFU_VERSION)

include boilerplate.mk

HOST_GOOS := $(shell go env GOHOSTOS)
HOST_GOARCH := $(shell go env GOHOSTARCH)
HOST_TOFU := build/.tools/tofu-$(TOFU_VERSION)-$(HOST_GOOS)_$(HOST_GOARCH)/tofu
comma := ,
MIRROR_CACHE := build/.mirror/$(subst $(comma),-,$(PACKAGED_CLOUDS))_$(GOOS)_$(GOARCH)
# Smallest release binary (MiB) that can hold OpenTofu plus one cloud's providers.
MIN_BINARY_MB ?= 100

.PHONY: deps build build-dev embed embed-clean host-tofu mirror-tofu fetch-tofu release check-release test fmt package-content
.PHONY: release-all release-azure release-aws release-gcp validate-examples e2e-sandbox e2e-azure-smoke e2e-azure-deploy deploy-unit deploy-e2e

validate-examples:
	@test -f $(DIST) || $(MAKE) build-dev
	@INSTALLER_CONTENT_ROOT=. ./$(DIST) validate examples/azure/sandbox.yaml --check-quota=false

e2e-sandbox:
	go test -tags=e2e -timeout=120m ./tests/e2e/ -run TestSandboxLifecycle

e2e-azure-smoke:
	go test -tags=e2e -timeout=30m ./tests/e2e/ -run 'TestAzureSmoke'

e2e-azure-deploy:
	go test -tags=e2e -timeout=180m ./tests/e2e/ -run 'TestAzureStackAction'

package-content:
	@PACKAGED_CLOUDS="$(PACKAGED_CLOUDS)" ./build/scripts/package-content.sh

deps:
	go mod download
	go mod tidy

fmt:
	go fmt ./...

build-dev: package-content
	@mkdir -p dist
	go build -ldflags "-X github.com/cipherops-io/rtifact-platform-installer/internal/version.PackagedClouds=$(PACKAGED_CLOUDS)" -o $(DIST) ./cmd/rtifact

embed-clean:
	@rm -rf build/embed/tofu build/embed/mirror

embed: fetch-tofu mirror-tofu
	@echo "embed artifacts ready under build/embed/ ($(GOOS)_$(GOARCH))"

fetch-tofu:
	@./build/scripts/fetch-tofu.sh $(TOFU_VERSION) $(GOOS) $(GOARCH) build/embed/tofu/$(GOOS)_$(GOARCH)

host-tofu: $(HOST_TOFU)

$(HOST_TOFU):
	@./build/scripts/fetch-tofu.sh $(TOFU_VERSION) $(HOST_GOOS) $(HOST_GOARCH) $(dir $@)

mirror-tofu: host-tofu
	@TOFU_BIN="$(abspath $(HOST_TOFU))" GOOS=$(GOOS) GOARCH=$(GOARCH) PACKAGED_CLOUDS="$(PACKAGED_CLOUDS)" \
	  MIRROR="$(MIRROR_CACHE)" ./build/scripts/mirror-providers.sh
	@mkdir -p build/embed/mirror && rsync -a "$(MIRROR_CACHE)/" build/embed/mirror/

# Order matters: the package is copied only after OpenTofu and the providers for GOOS/GOARCH are in build/embed.
release:
	@$(MAKE) --no-print-directory embed-clean
	@$(MAKE) --no-print-directory embed
	@$(MAKE) --no-print-directory package-content
	@mkdir -p dist
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -trimpath -ldflags="$(LDFLAGS)" -o $(DIST) ./cmd/rtifact
	@$(MAKE) --no-print-directory check-release

check-release:
	@./build/scripts/check-release.sh $(DIST) $(GOOS) $(GOARCH) "$(PACKAGED_CLOUDS)" $(MIN_BINARY_MB)

# Push <PUBLISH_DIR>/rtifact-<cloud>-<goos>-<goarch> to the registry (oras login first).
# REPLACE_EXISTING, SIGN, COSIGN_PRIVATE_KEY, COSIGN_PASSWORD, SOURCE_REPO, SUMMARY_FILE come from the environment.
PUBLISH_DIR ?= out
REGISTRY ?= zot.prod.rtifact.ai
REGISTRY_REPOSITORY ?= rtifact-installer
PUBLISH_PLATFORMS ?= linux/amd64,linux/arm64,darwin/arm64,darwin/amd64

.PHONY: publish
publish:
	@DIST_DIR="$(PUBLISH_DIR)" REGISTRY="$(REGISTRY)" REPOSITORY="$(REGISTRY_REPOSITORY)" VERSION="$(VERSION)" \
	  CLOUDS="$(PACKAGED_CLOUDS)" PLATFORMS="$(PUBLISH_PLATFORMS)" PRODUCT="$(BINARY)" \
	  GIT_COMMIT="$(GIT_COMMIT)" TOFU_VERSION="$(TOFU_VERSION)" ./build/scripts/publish-zot.sh

release-all:
	$(MAKE) release PACKAGED_CLOUDS=azure,aws,gcp

release-azure:
	$(MAKE) release PACKAGED_CLOUDS=azure

release-aws:
	$(MAKE) release PACKAGED_CLOUDS=aws

release-gcp:
	$(MAKE) release PACKAGED_CLOUDS=gcp

# No cluster: fakes for Kubernetes and Helm. Cluster scenarios are deploy-e2e.
deploy-unit:
	go test ./internal/deploy/... ./internal/archtest/... ./internal/cli/... ./tests/deploy/...

deploy-e2e:
	@chmod +x build/scripts/deploy-e2e.sh tests/deploy/rehearsal.sh 2>/dev/null || true
	@./build/scripts/deploy-e2e.sh

VERSION_CLOUDS := azure aws gcp
VERSION_LAYERS := state-backend core-network data-platform app-platform network-edge ai-integrations

.PHONY: bump-providers
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
	      --var-file build/provider-versions.yml \
	      --var "StackTarget=$$target" \
	      --var "OpenTofuVersion=$(TOFU_VERSION)" \
	      --var 'ProviderSources={azure: "$(PROVIDER_SOURCE_AZURE)", aws: "$(PROVIDER_SOURCE_AWS)", gcp: "$(PROVIDER_SOURCE_GCP)"}'; \
	  done; \
	done
	@$(MAKE) tofu-lock

test:
	go test ./...

tofu-stack-terratest:
	@cd tests/stacks/terratest && go test -count=1 ./...