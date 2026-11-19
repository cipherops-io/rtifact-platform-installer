# AGENTS.md — rtifact-platform-installer

Instructions for humans and AI agents working in this repository.

## Runtime

- **OpenTofu only.** Use the `tofu` CLI (`plan`, `apply`, `destroy`, `validate`, `fmt`, `init`). Never invoke HashiCorp Terraform (`terraform`).
- The Go installer (`cmd/installer`) only executes `tofu`.

## Provider sources

- Pin providers from the **[OpenTofu Registry](https://search.opentofu.org/providers)** only.
- Use the **`opentofu/` namespace** (e.g. `opentofu/azurerm`, `opentofu/random`, `opentofu/tls`), not `hashicorp/*`.
- Versions are declared in each module/stack `providers.tofu` inside `terraform { required_providers { ... } }` (OpenTofu settings block syntax).
- On `tofu init`, the registry may show informational notes comparing `opentofu/*` and `hashicorp/*` namespaces; **ignore those** — this project standardizes on `opentofu/*`.

## Configuration files

- **Do not add new `.tf` files.** All OpenTofu configuration must use the **`.tofu`** extension (e.g. `main.tofu`, `variables.tofu`, `providers.tofu`).
- Variable files for customers stay **YAML** (e.g. `<customer>-<env>-<purpose>.yaml` **external** to the repo).
- Customer contract (schema v2): top-level `cloud`, `account`, `workload`, and `features.*` booleans (validated against the profile registry).
- Generated at apply time: `installer.auto.tfvars.json` (passed via `-var-file`) and `backend.tofu` (from `cloud.Backend()`); local state lives under `~/.rtifact-installer/<prefix>/state/`.
- New IaC modules/resources must be behind a `features.<name>` flag in the stack.

## Layout

- `infra/modules/<cloud>/` — reusable building blocks (`providers.tofu` + `*.tofu`).
- `infra/stacks/<profile>/<cloud>/<layer>/` — composition roots; **only stacks** configure `provider "azurerm"` blocks.
- `profiles/` — workload defaults (not customer-specific).
- Customer values live **outside** this repo.

## When changing IaC

1. Match [azurerm 5.x](https://search.opentofu.org/provider/opentofu/azurerm/latest) schemas; avoid deprecated arguments (e.g. storage account uses `public_network_access`, not `public_network_access_enabled`).
2. Run `tofu fmt -recursive infra/` and `make tofu-validate`.
3. Update `infra/stacks/<profile>/<cloud>/<layer>/.terraform.lock.hcl` when provider versions change (`tofu init -upgrade` in the stack dir).
4. Run `make package-content` before release builds.

## Build

```bash
make build-dev
make tofu-validate
./dist/rtifact validate examples/sandbox.yaml
```

Release: `make release` (embeds OpenTofu + provider mirror).