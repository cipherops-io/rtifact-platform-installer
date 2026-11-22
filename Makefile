MODULE := github.com/cipherops-io/rtifact-platform-installer
BINARY := rtifact
DIST := dist/$(BINARY)
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
PACKAGED_CLOUDS ?= azure,aws,gcp
VERSION ?= 0.0.0-dev
CUSTOMER ?= internal
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
LDFLAGS := -s -w \
  -X github.com/cipherops-io/rtifact-platform-installer/internal/version.Release=true \
  -X github.com/cipherops-io/rtifact-platform-installer/internal/version.PackagedClouds=$(PACKAGED_CLOUDS) \
  -X github.com/cipherops-io/rtifact-platform-installer/internal/version.Version=$(VERSION) \
  -X github.com/cipherops-io/rtifact-platform-installer/internal/version.GitCommit=$(GIT_COMMIT) \
  -X github.com/cipherops-io/rtifact-platform-installer/internal/version.Customer=$(CUSTOMER)

include boilerplate.mk

.PHONY: deps build build-dev embed mirror-tofu fetch-tofu release test fmt package-content
.PHONY: release-all release-azure release-aws release-gcp validate-examples e2e-sandbox e2e-azure-smoke e2e-azure-deploy deploy-e2e

validate-examples:
	@test -f $(DIST) || $(MAKE) build-dev
	@INSTALLER_CONTENT_ROOT=. ./$(DIST) validate examples/azure/sandbox.yaml --check-azure=false

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
	go build -ldflags "-X github.com/cipherops-io/rtifact-platform-installer/internal/version.PackagedClouds=$(PACKAGED_CLOUDS)" -o $(DIST) ./cmd/installer

embed: fetch-tofu mirror-tofu
	@echo "embed artifacts ready under build/embed/"

fetch-tofu:
	@./build/scripts/fetch-tofu.sh $(TOFU_VERSION) $(GOOS) $(GOARCH)

mirror-tofu:
	@TOFU_BIN="$(TOFU_BIN)" PACKAGED_CLOUDS="$(PACKAGED_CLOUDS)" ./build/scripts/mirror-providers.sh

release: package-content embed
	@mkdir -p dist
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="$(LDFLAGS)" -o $(DIST) ./cmd/installer

release-all:
	$(MAKE) release PACKAGED_CLOUDS=azure,aws,gcp

release-azure:
	$(MAKE) release PACKAGED_CLOUDS=azure

release-aws:
	$(MAKE) release PACKAGED_CLOUDS=aws

release-gcp:
	$(MAKE) release PACKAGED_CLOUDS=gcp

deploy-e2e:
	go test ./internal/deploy/... ./tests/deploy/...

test:
	go test ./...

tofu-stack-terratest:
	@cd tests/stacks/terratest && go test -count=1 ./...