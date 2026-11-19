output "communication_service_id" {
  value = length(azurerm_communication_service.email) > 0 ? azurerm_communication_service.email[0].id : null
}

output "email_communication_service_id" {
  value = length(azurerm_email_communication_service.main) > 0 ? azurerm_email_communication_service.main[0].id : null
}

output "ai_foundry_id" {
  value = length(module.ai_foundry) > 0 ? module.ai_foundry[0].ai_foundry_id : null
}