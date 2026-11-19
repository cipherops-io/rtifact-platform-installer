output "application_gateway_id" {
  value = length(module.app_gateway) > 0 ? module.app_gateway[0].application_gateway_id : null
}

output "public_dns_zone_ids" {
  value = length(module.public_dns) > 0 ? module.public_dns[0].zone_ids : {}
}