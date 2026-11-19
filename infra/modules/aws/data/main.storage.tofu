resource "aws_s3_bucket" "main" {
  count = var.features.storage_account ? 1 : 0

  bucket = "${substr(replace(lower(var.name_prefix), "_", "-"), 0, 40)}-${random_string.suffix[0].result}"
  tags   = var.tags
}

resource "aws_s3_bucket_public_access_block" "main" {
  count = var.features.storage_account ? 1 : 0

  bucket                  = aws_s3_bucket.main[0].id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}