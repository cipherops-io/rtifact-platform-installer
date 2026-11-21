module "state_backend" {
  source = "../../../../modules/aws/state-backend"

  name_prefix = var.name_prefix
  location    = var.location
  account_id  = var.account_id
  state       = try(var.state, {})
}