resource "aws_kms_key" "main" {
  description             = "${var.name_prefix} workload encryption key"
  deletion_window_in_days = 7
  enable_key_rotation     = true
  tags                    = var.tags
}

resource "aws_kms_alias" "main" {
  name          = "alias/${replace(var.name_prefix, "-", "_")}"
  target_key_id = aws_kms_key.main.key_id
}
