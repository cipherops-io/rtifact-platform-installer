# Remote state storage in the same resource group as the workload ({prefix}-rg).
# Applied with local state by `installer bootstrap`.
module "state_backend" {
  source = "../../../../modules/azure/state-backend"

  name_prefix           = var.name_prefix
  location              = var.location
  subscription_id       = var.subscription_id
  resource_group_name   = var.resource_group_name
  create_resource_group = var.resource_group_create
}
