# Placeholder layer: resources are added here later.
data "terraform_remote_state" "core_network" {
  backend = var.upstream_state["core-network"].backend
  config  = var.upstream_state["core-network"].config
}

locals {
  core = data.terraform_remote_state.core_network.outputs
}
