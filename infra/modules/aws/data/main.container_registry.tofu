resource "aws_ecr_repository" "main" {
  count = var.features.container_registry ? 1 : 0

  name                 = "${local.base_name}-${random_string.suffix[0].result}"
  image_tag_mutability = var.container_registry.image_tag_mutability
  tags                 = var.tags

  image_scanning_configuration {
    scan_on_push = var.container_registry.scan_on_push
  }

  encryption_configuration {
    encryption_type = var.kms_key_arn == null ? "AES256" : "KMS"
    kms_key         = var.kms_key_arn
  }
}
