data "terraform_remote_state" "core_network" {
  backend = var.upstream_state["core-network"].backend
  config  = var.upstream_state["core-network"].config
}

locals {
  core = data.terraform_remote_state.core_network.outputs
}

locals {
  data_features = {
    postgresql         = try(var.features.postgresql, false)
    redis              = try(var.features.redis, false)
    object_storage     = try(var.features.object_storage, false)
    container_registry = try(var.features.container_registry, false)
  }
  data_services_enabled = anytrue(values(local.data_features))
}

module "security" {
  count = try(var.features.kms, false) ? 1 : 0

  source                  = "../../../../modules/aws/security"
  name_prefix             = var.name_prefix
  tags                    = local.core.tags
  deletion_window_days    = var.kms.deletion_window_days
  enable_key_rotation     = var.kms.enable_key_rotation
  rotation_period_in_days = var.kms.rotation_period_in_days
}

module "data" {
  count = local.data_services_enabled ? 1 : 0

  source             = "../../../../modules/aws/data"
  name_prefix        = var.name_prefix
  tags               = local.core.tags
  features           = local.data_features
  vpc_id             = local.core.vpc_id
  vpc_cidrs          = local.core.vpc_cidrs
  subnet_ids_by_tier = local.core.subnet_ids_by_tier
  kms_key_arn        = one(module.security[*].kms_key_arn)
  postgresql         = var.postgresql
  redis              = var.redis
  object_storage     = var.object_storage
  container_registry = var.container_registry
}
