# azure-platform-installer

Customer-agnostic Azure infrastructure installer: a Go binary that embeds OpenTofu, provider mirrors, workload profiles, and composable IaC stacks.

**Contributors / agents:** read [AGENTS.md](AGENTS.md) first.

Customers supply an external `config.yaml` (never committed to this repo). See [docs/customer-runbook.md](docs/customer-runbook.md).

## Quick start (developers)

Requires the **OpenTofu** CLI (`tofu` on `PATH`, or embedded in release builds). IaC uses **`.tofu`** files only (see `AGENTS.md`), `default.auto.tofuvars.json`, and lock files; providers come from the [OpenTofu Registry](https://search.opentofu.org/providers) only (`opentofu/azurerm`, `opentofu/random`, `opentofu/tls`). Only the `opentofu/*` provider namespace on registry.opentofu.org.

```bash
make deps
make build-dev
./dist/installer validate examples/ai-os-centralindia/config.yaml
```

## Release build

```bash
make embed        # fetch tofu + mirror providers (requires network)
make release
```

## Layout

- `cmd/installer` — CLI entrypoint
- `profiles/` — workload defaults (e.g. `ai-os-aks`)
- `infra/modules/` — reusable OpenTofu modules
- `infra/stacks/` — composition roots (providers only in stack)
- `build/embed/` — populated before release (gitignored)