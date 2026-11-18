module "state_backend" {
  source = "../../../modules/state-backend"
  name_prefix         = var.name_prefix
  location            = var.location
  subscription_id     = var.subscription_id
  resource_group_name = "${var.name_prefix}-tfstate-rg"
}