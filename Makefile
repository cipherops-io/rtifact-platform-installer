MODULE := github.com/cipherops-io/rtifact-platform-installer
BINARY := installer
DIST := dist/$(BINARY)
TOFU_VERSION ?= 1.9.0
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

.PHONY: deps build build-dev embed mirror-tofu fetch-tofu release test fmt package-content tofu-fmt tofu-validate

tofu-fmt:
	tofu fmt -recursive infra/

tofu-validate:
	@for d in infra/stacks/ai-os-aks infra/stacks/ai-os-aks/network infra/stacks/ai-os-aks/data infra/stacks/ai-os-aks/cluster infra/stacks/ai-os-aks/bootstrap; do \
	  [ -f "$$d/providers.tofu" ] && (cd "$$d" && tofu init -backend=false -input=false && tofu validate) || true; \
	done

validate-examples:
	@test -f dist/installer || $(MAKE) build-dev
	@INSTALLER_CONTENT_ROOT=. ./dist/installer validate examples/sandbox.yaml --check-azure=false

.PHONY: validate-examples e2e-sandbox
e2e-sandbox:
	go test -tags=e2e -timeout=120m ./tests/e2e/...

package-content:
	@./build/scripts/package-content.sh

deps:
	go mod download
	go mod tidy

fmt:
	go fmt ./...

build-dev: package-content
	@mkdir -p dist
	go build -o $(DIST) ./cmd/installer

embed: fetch-tofu mirror-tofu
	@echo "embed artifacts ready under build/embed/"

fetch-tofu:
	@./build/scripts/fetch-tofu.sh $(TOFU_VERSION) $(GOOS) $(GOARCH)

mirror-tofu:
	@TOFU_BIN="$(TOFU_BIN)" ./build/scripts/mirror-providers.sh

release: package-content embed
	@mkdir -p dist
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="-s -w -X github.com/cipherops-io/rtifact-platform-installer/internal/version.Release=true" -o $(DIST) ./cmd/installer

test:
	go test ./...