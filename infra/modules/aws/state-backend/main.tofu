locals {
  bucket_name       = substr(replace(lower(var.name_prefix), "_", "-"), 0, 45)
  bucket_name_final = "${local.bucket_name}-tfstate"
  lock_table_name   = substr(replace(var.name_prefix, "-", "_"), 0, 200)
  lock_table_final  = "${local.lock_table_name}_tfstate_lock"
}

resource "aws_s3_bucket" "tfstate" {
  bucket = local.bucket_name_final
  tags = {
    managed-by = "rtifact"
    purpose    = "terraform-state"
  }
}

resource "aws_s3_bucket_versioning" "tfstate" {
  bucket = aws_s3_bucket.tfstate.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "tfstate" {
  bucket = aws_s3_bucket.tfstate.id
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
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
  name         = local.lock_table_final
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "LockID"

  attribute {
    name = "LockID"
    type = "S"
  }

  tags = {
    managed-by = "rtifact"
    purpose    = "terraform-state-lock"
  }
}
