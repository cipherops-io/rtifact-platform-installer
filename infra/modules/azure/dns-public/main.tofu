resource "azurerm_dns_zone" "zone" {
  for_each = { for z in var.zones : z.name => z }

  name                = each.value.name
  resource_group_name = var.resource_group_name
  tags                = var.tags
}

resource "azurerm_dns_a_record" "record" {
  for_each = {
    for pair in flatten([
      for z in var.zones : [
        for r in try(z.records, []) : {
          key   = "${z.name}/${r.name}/${r.type}"
          zone  = z.name
          name  = r.name
          type  = r.type
          ttl   = try(r.ttl, 300)
          value = r.record
        } if r.type == "A"
      ]
    ]) : pair.key => pair
  }

  name                = each.value.name
  zone_name           = azurerm_dns_zone.zone[each.value.zone].name
  resource_group_name = var.resource_group_name
  ttl                 = each.value.ttl
  records             = [each.value.value]
}