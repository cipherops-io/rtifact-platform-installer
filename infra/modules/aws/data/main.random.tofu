locals {
  base_name        = trim(substr(replace(lower(var.name_prefix), "_", "-"), 0, 34), "-")
  need_name_suffix = var.features.object_storage || var.features.container_registry
}

resource "random_string" "suffix" {
  count = local.need_name_suffix ? 1 : 0

  length  = 6
  lower   = true
  upper   = false
  special = false
}
