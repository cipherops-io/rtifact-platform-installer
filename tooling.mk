# DO NOT EDIT — managed by cipherops-io/boilerplates (rtifact-product).
# Change it there, then run: make boilerplate-update

# Tool and CLI versions — rendered from rtifact-product defaults.

TOFU_VERSION ?= 1.13.9
TERRAFORM_DOCS_VERSION ?= 0.20.0
BOILERPLATE_VERSION ?= 0.17.0
AZ_CLI_VERSION ?= 2.93.0
AWS_CLI_VERSION ?= 2.41.1
GCP_CLI_VERSION ?= 502.0.0
PROVIDER_SOURCE_AZURE ?= opentofu/azurerm
PROVIDER_SOURCE_AWS ?= opentofu/aws
PROVIDER_SOURCE_GCP ?= opentofu/google
LOCK_PLATFORMS ?= linux_amd64,darwin_arm64,darwin_amd64,windows_amd64

.PHONY: install-boilerplate-cli

install-boilerplate-cli:
	@set -e; \
	os=$$(uname -s | tr '[:upper:]' '[:lower:]'); \
	arch=$$(uname -m); case "$$arch" in x86_64|amd64) arch=amd64;; aarch64|arm64) arch=arm64;; *) echo "unsupported arch"; exit 1;; esac; \
	dest="$${HOME}/.local/bin"; mkdir -p "$$dest"; \
	curl -fsSL -o "$$dest/boilerplate" \
	  "https://github.com/gruntwork-io/boilerplate/releases/download/v$(BOILERPLATE_VERSION)/boilerplate_$${os}_$${arch}"; \
	chmod +x "$$dest/boilerplate"; \
	"$$dest/boilerplate" --version