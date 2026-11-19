module "state_backend" {
  source = "../../../../modules/gcp/state-backend"

  name_prefix = var.name_prefix
  location    = var.location
  project_id  = var.project_id
}
