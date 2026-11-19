locals {
  key_vault_pe_enabled = var.features.key_vault_private_endpoint && var.features.key_vault
}

module "kms_pe" {
  count = local.key_vault_pe_enabled ? 1 : 0

  source       = "../../../../modules/aws/private-endpoint"
  name         = "${var.name_prefix}-kms-pe"
  location     = var.location
  scope_name   = local.core.resource_group_name
  subnet_id    = local.core.private_endpoint_subnet_id
  vpc_id       = local.core.vnet_id
  service_name = "com.amazonaws.${var.location}.kms"
  tags         = local.core.tags

  depends_on = [module.security]
}
