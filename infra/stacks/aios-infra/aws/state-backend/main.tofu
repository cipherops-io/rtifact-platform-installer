# Remote state storage for all downstream layers. Applied with local state by installer bootstrap.
module "state_backend" {
  source = "../../../../modules/aws/state-backend"

  name_prefix = var.name_prefix
  location    = var.location
  account_id  = var.account_id
}
