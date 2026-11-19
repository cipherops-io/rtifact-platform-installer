locals {
  scope_name = "${var.prefix}-rg"
  tags = merge({
    "managed-by" = "rtifact"
    "scope"      = var.prefix
    "location"   = var.location
  }, var.tags)
}
