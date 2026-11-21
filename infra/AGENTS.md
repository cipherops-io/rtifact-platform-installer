<!-- DO NOT EDIT — managed by cipherops-io/boilerplates (rtifact-product).
     Change it there, then run: make boilerplate-update -->
# AGENTS.md — infra

- **Modules** `infra/modules/<cloud>/<name>/` — reusable building blocks.
- **Stacks** `infra/stacks/<profile>/<cloud>/<layer>/` — composition roots; only stacks configure cloud `provider` blocks.
- **Profiles** `profiles/<profile>/<cloud>/` — defaults, features, tiers for one cloud only.
- Edit **`.tofu` logic here**; do not hand-edit managed `AGENTS.md` files — update boilerplates and `make boilerplate-update`.

## Cloud layout rules

1. Keep each cloud's IaC and profile data under that cloud's folders only; never import or name another cloud's resources in the wrong tree.
2. Gate capabilities with a single `features.<name>` flag (default false); remove `enabled` sub-keys from YAML sections.
3. Do not encode customer defaults in `variables.tofu` or other `*.tofu` files — use `profiles/<id>/<cloud>/defaults.yaml`.

## How to place a new service {#placement}

1. **Look first.** Search `infra/` and `profiles/` for the service or module name. If it exists, use its layer.
2. **Feature flag.** `grep -rl 'features\.<flag>' infra/stacks/` shows the layer that uses it (authoritative).
3. **New service.** Identify what kind of service it is (cloud docs or web search for classification only), then answer **in order** — first yes wins:
   1. Stores this installer's own remote state? → `state-backend`
   2. Foundation other layers build on (network, subnets, NAT, egress firewall, jump host, VPN/ExpressRoute)? → `core-network`
   3. Holds data, secrets, keys, images or files, or private endpoint/DNS for such a service? → `data-platform`
   4. Runs workloads (Kubernetes, node pools, VMs, workload identity)? → `app-platform`
   5. Faces the public internet (gateway/LB+WAF, public DNS, edge TLS)? → `network-edge`
   6. Managed AI model access or transactional email? → `ai-integrations`
   7. None match, or two match equally → **stop and ask** — do not guess.
4. **Tie-breakers.** Private endpoint follows its data service (`data-platform`). Egress firewall is `core-network`. Model hub is `ai-integrations`; key vault and storage it needs are `data-platform`.
5. **Cross-cloud.** Same layer on all clouds; only the module folder name differs (`aks` / `eks` / `gke`). Create under `infra/modules/<cloud>/` and wire in the same layer.
6. **Finish.** Gate with `features.<name>`, register `layer/module` in `boilerplate-vars.yml`, run `make tofu-pre-commit-fast`.
7. **Never add logic to boilerplates** — resources and logic live in this installer repo only.

Per-layer summaries: `infra/stacks/<profile>/<cloud>/<layer>/AGENTS.md`. Output contracts: `docs/layer-contracts.md`.