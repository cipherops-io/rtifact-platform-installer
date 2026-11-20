<!-- DO NOT EDIT — managed by cipherops-io/boilerplates (rtifact-product).
     Change it there, then run: make boilerplate-update -->
# AGENTS.md — infra

- **Modules** `infra/modules/<cloud>/<name>/` — reusable building blocks.
- **Stacks** `infra/stacks/<profile>/<cloud>/<layer>/` — composition roots only stacks configure cloud `provider` blocks.
- Edit **`.tofu` logic here**; do not change managed `AGENTS.md` files by hand — update boilerplates and `make boilerplate-update`.