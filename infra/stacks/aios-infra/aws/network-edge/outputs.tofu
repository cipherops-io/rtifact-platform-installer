output "dns_zone_id" {
  value = length(module.dns_public) > 0 ? module.dns_public[0].zone_id : null
}

output "dns_name_servers" {
  value = length(module.dns_public) > 0 ? module.dns_public[0].name_servers : []
}

output "certificate_regional_arn" {
  value = length(module.certificates) > 0 ? module.certificates[0].regional_certificate_arn : null
}