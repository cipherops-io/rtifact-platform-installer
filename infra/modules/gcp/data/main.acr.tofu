resource "google_artifact_registry_repository" "main" {
  count = var.features.container_registry ? 1 : 0

  location      = var.location
  repository_id = "${substr(replace(lower(var.name_prefix), "_", "-"), 0, 40)}-${random_string.suffix[0].result}"
  format        = "DOCKER"
}