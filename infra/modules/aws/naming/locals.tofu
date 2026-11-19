locals {
  caf_prefixes = {
    resource_group = "rg"
    vpc            = "vpc"
    kms            = "kms"
    eks            = "eks"
    s3             = "s3"
  }
  tags = merge({
    "managed-by" = "rtifact"
    "scope"      = var.prefix
    "location"   = var.location
  }, var.tags)
}
