# Remote state storage for all downstream layers. Applied with local state by
# `installer bootstrap`. Names must match internal/cloud/azure (StateResourceGroupName /
# StateStorageAccountName).
module "state_backend" {
  source = "../../../../modules/azure/state-backend"

  name_prefix         = var.name_prefix
  location            = var.location
  subscription_id     = var.subscription_id
  resource_group_name = "${var.name_prefix}-tfstate-rg"
}
