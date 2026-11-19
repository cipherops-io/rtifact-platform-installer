locals {
  tags = merge({
    "managed-by" = "rtifact"
  }, var.tags)
}