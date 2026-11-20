data "terraform_remote_state" "data_platform" {
  backend = var.upstream_state["data-platform"].backend
  config  = var.upstream_state["data-platform"].config
}

locals {
  data_outputs                  = try(data.terraform_remote_state.data_platform.outputs, {})
  ai_foundry_storage_account_id = try(local.data_outputs.ai_foundry_storage_account_id, null)
  ai_foundry_key_vault_id       = try(local.data_outputs.ai_foundry_key_vault_id, null)
  acr_id                        = try(local.data_outputs.acr_id, null)
  foundry_upstream_ready = (
    local.ai_foundry_storage_account_id != null &&
    local.ai_foundry_key_vault_id != null
  )
}

module "ai_foundry" {
  count = var.features.ai_foundry && local.foundry_upstream_ready ? 1 : 0

  source                = "../../../../modules/azure/ai-foundry"
  resource_group_name   = local.core.resource_group_name
  location              = var.location
  name_prefix           = var.name_prefix
  storage_account_id    = local.ai_foundry_storage_account_id
  key_vault_id          = local.ai_foundry_key_vault_id
  container_registry_id = local.acr_id
  public_network_access = var.ai_foundry.public_network_access
  tags                  = local.core.tags
}