resource "azurerm_communication_service" "email" {
  count = var.features.communication_email ? 1 : 0

  name                = "${var.name_prefix}-acs"
  resource_group_name = local.core.resource_group_name
  data_location       = try(var.communication_services.data_location, "India")
  tags                = local.core.tags
}

resource "azurerm_email_communication_service" "main" {
  count = var.features.communication_email ? 1 : 0

  name                = "${var.name_prefix}-ecs"
  resource_group_name = local.core.resource_group_name
  data_location       = try(var.communication_services.data_location, "India")
  tags                = local.core.tags
}