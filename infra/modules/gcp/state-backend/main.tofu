locals {
  bucket_name = substr(replace(lower(var.name_prefix), "_", "-"), 0, 50)
}

resource "google_storage_bucket" "tfstate" {
  name                        = "${local.bucket_name}-tfstate"
  location                    = upper(var.location)
  force_destroy               = false
  uniform_bucket_level_access = true
  versioning { enabled = true }

  labels = {
    managed-by = "rtifact"
    purpose    = "terraform-state"
  }
}
