resource "aws_ecr_repository" "main" {
  count = var.features.container_registry ? 1 : 0

  name                 = "${substr(replace(lower(var.name_prefix), "_", "-"), 0, 40)}-${random_string.suffix[0].result}"
  image_tag_mutability = "IMMUTABLE"
  tags                 = var.tags
}