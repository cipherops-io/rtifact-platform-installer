output "communication_service_id" {
  value = length(azurerm_communication_service.email) > 0 ? azurerm_communication_service.email[0].id : null
}

output "email_communication_service_id" {
  value = length(azurerm_email_communication_service.main) > 0 ? azurerm_email_communication_service.main[0].id : null
}