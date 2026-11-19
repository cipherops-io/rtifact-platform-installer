# Workload profiles

Profiles ship **defaults only** (`features`, sizing, network). Customers use an external YAML file with `profile: <id>` and top-level sections (`features`, `kubernetes_cluster`, etc.).

Add a profile by copying `ai-os-k8s/` and pointing `stack_id` at an `infra/stacks/<id>/<cloud>/<layer>` composition (layers: state-backend, core-network, data-platform, app-platform, network-edge, ai-integrations).