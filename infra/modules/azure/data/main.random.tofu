locals {
  need_name_suffix = var.features.storage_account || var.features.container_registry
}

resource "random_string" "suffix" {
  count = local.need_name_suffix ? 1 : 0

  length  = 6
  lower   = true
  upper   = false
  special = false
}