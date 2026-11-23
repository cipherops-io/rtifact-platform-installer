resource "aws_kms_key" "main" {
  description             = "${var.name_prefix} data services encryption key"
  deletion_window_in_days = var.deletion_window_days
  enable_key_rotation     = var.enable_key_rotation
  rotation_period_in_days = var.rotation_period_in_days
  tags                    = var.tags
}

resource "aws_kms_alias" "main" {
  name          = "alias/${replace(var.name_prefix, "-", "_")}"
  target_key_id = aws_kms_key.main.key_id
}
