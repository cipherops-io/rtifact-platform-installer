resource "azurerm_resource_group" "main" {
  count = var.create_resource_group ? 1 : 0

  name     = "${var.prefix}-rg"
  location = var.location
  tags     = local.tags
}

data "azurerm_resource_group" "main" {
  count = var.create_resource_group ? 0 : 1

  name = "${var.prefix}-rg"
}

locals {
  resource_group_name = var.create_resource_group ? azurerm_resource_group.main[0].name : data.azurerm_resource_group.main[0].name
  resource_group_id   = var.create_resource_group ? azurerm_resource_group.main[0].id : data.azurerm_resource_group.main[0].id
  effective_tags      = var.create_resource_group ? local.tags : data.azurerm_resource_group.main[0].tags
}