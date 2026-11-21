data "terraform_remote_state" "core_network" {
  backend = var.upstream_state["core-network"].backend
  config  = var.upstream_state["core-network"].config
}

locals {
  base_tags = try(var.tags, {})
  _core     = data.terraform_remote_state.core_network.outputs
}

module "dns_public" {
  count = var.features.public_dns ? 1 : 0

  source      = "../../../../modules/aws/dns-public"
  name_prefix = var.name_prefix
  tags        = local.base_tags
  public_dns  = var.public_dns
}

module "certificates" {
  count = var.features.certificates && length(module.dns_public) > 0 && (try(var.dns_delegation_verified, false) || try(var.public_dns.delegated, false) || try(var.public_dns.parent_zone_id, "") != "") ? 1 : 0

  source       = "../../../../modules/aws/certificates"
  name_prefix  = var.name_prefix
  location     = var.location
  tags         = local.base_tags
  certificates = var.certificates
  zone_id      = module.dns_public[0].zone_id

  providers = {
    aws.us_east_1 = aws.us_east_1
  }

  depends_on = [module.dns_public]
}