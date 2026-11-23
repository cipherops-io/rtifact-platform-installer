data "terraform_remote_state" "core_network" {
  backend = var.upstream_state["core-network"].backend
  config  = var.upstream_state["core-network"].config
}

locals {
  core = data.terraform_remote_state.core_network.outputs
}

module "app_gateway" {
  count = var.features.application_gateway ? 1 : 0

  source                = "../../../../modules/azure/app-gateway"
  resource_group_name   = local.core.resource_group_name
  location              = var.location
  name_prefix           = var.name_prefix
  subnet_id             = local.core.load_balancer_subnet_id
  sku_name              = var.application_gateway.sku_name
  sku_tier              = var.application_gateway.sku_tier
  capacity              = var.application_gateway.capacity
  ssl_policy            = var.application_gateway.ssl_policy
  frontend_port         = var.application_gateway.frontend_port
  backend_http_settings = var.application_gateway.backend_http_settings
  routing_rule_priority = var.application_gateway.routing_rule_priority
  tags                  = local.core.tags
}

module "public_dns" {
  count = var.features.public_dns_zone && length(var.public_dns.zones) > 0 ? 1 : 0

  source              = "../../../../modules/azure/dns-public"
  resource_group_name = local.core.resource_group_name
  zones               = var.public_dns.zones
  tags                = local.core.tags
}