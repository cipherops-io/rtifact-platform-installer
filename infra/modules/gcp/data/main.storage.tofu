resource "google_storage_bucket" "main" {
  count = var.features.storage_account ? 1 : 0

  name                        = "${substr(replace(lower(var.name_prefix), "_", "-"), 0, 40)}-${random_string.suffix[0].result}"
  location                    = upper(var.location)
  force_destroy               = false
  uniform_bucket_level_access = true
}