locals {
  tags = merge({
    "managed-by" = "azure-platform-installer"
  }, var.tags)
}