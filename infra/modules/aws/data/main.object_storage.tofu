resource "aws_s3_bucket" "main" {
  count = var.features.object_storage ? 1 : 0

  bucket = "${local.base_name}-${random_string.suffix[0].result}"
  tags   = var.tags
}

resource "aws_s3_bucket_ownership_controls" "main" {
  count = var.features.object_storage ? 1 : 0

  bucket = aws_s3_bucket.main[0].id
  rule {
    object_ownership = "BucketOwnerEnforced"
  }
}

resource "aws_s3_bucket_public_access_block" "main" {
  count = var.features.object_storage ? 1 : 0

  bucket                  = aws_s3_bucket.main[0].id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_versioning" "main" {
  count = var.features.object_storage ? 1 : 0

  bucket = aws_s3_bucket.main[0].id
  versioning_configuration {
    status = var.object_storage.versioning ? "Enabled" : "Suspended"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "main" {
  count = var.features.object_storage ? 1 : 0

  bucket = aws_s3_bucket.main[0].id
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm     = var.kms_key_arn == null ? "AES256" : "aws:kms"
      kms_master_key_id = var.kms_key_arn
    }
    bucket_key_enabled = var.kms_key_arn != null
  }
}
