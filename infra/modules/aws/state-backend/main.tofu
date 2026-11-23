locals {
  bucket_name = coalesce(
    try(var.state.bucket_name, null),
    "${substr(replace(lower(var.name_prefix), "_", "-"), 0, 45)}-tfstate-${var.account_id}",
  )
  lock_table_name = substr(replace(var.name_prefix, "-", "_"), 0, 200)
  lock_table      = "${local.lock_table_name}_tfstate_lock"
  versioning      = var.state.versioning
  sse_algorithm   = var.state.encryption
}

resource "aws_s3_bucket" "tfstate" {
  bucket = local.bucket_name
  tags = {
    managed-by = "rtifact"
    purpose    = "terraform-state"
  }
}

resource "aws_s3_bucket_versioning" "tfstate" {
  bucket = aws_s3_bucket.tfstate.id
  versioning_configuration {
    status = local.versioning ? "Enabled" : "Suspended"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "tfstate" {
  bucket = aws_s3_bucket.tfstate.id
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = local.sse_algorithm
    }
  }
}

resource "aws_s3_bucket_public_access_block" "tfstate" {
  bucket                  = aws_s3_bucket.tfstate.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_dynamodb_table" "tfstate_lock" {
  name           = local.lock_table
  billing_mode   = var.state.lock_table.billing_mode
  read_capacity  = var.state.lock_table.billing_mode == "PROVISIONED" ? var.state.lock_table.read_capacity : null
  write_capacity = var.state.lock_table.billing_mode == "PROVISIONED" ? var.state.lock_table.write_capacity : null
  hash_key       = "LockID"

  attribute {
    name = "LockID"
    type = "S"
  }

  tags = {
    managed-by = "rtifact"
    purpose    = "terraform-state-lock"
  }
}