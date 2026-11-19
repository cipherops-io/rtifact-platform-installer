data "terraform_remote_state" "data_platform" {
  backend = var.upstream_state["data-platform"].backend
  config  = var.upstream_state["data-platform"].config
}

locals {
  data = data.terraform_remote_state.data_platform.outputs
}

module "ai_foundry" {
  count = var.features.ai_foundry ? 1 : 0

  source                = "../../../../modules/azure/ai-foundry"
  resource_group_name   = local.core.resource_group_name
  location              = var.location
  name_prefix           = var.name_prefix
  storage_account_id    = local.data.storage_account_id
  key_vault_id          = local.data.key_vault_id
  container_registry_id = local.data.acr_id
  public_network_access = var.ai_foundry.public_network_access
  tags                  = local.core.tags
}