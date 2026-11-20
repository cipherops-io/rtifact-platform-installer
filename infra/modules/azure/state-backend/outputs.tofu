output "storage_account_name" {
  value = var.create_storage_account ? azurerm_storage_account.tfstate[0].name : data.azurerm_storage_account.tfstate[0].name
}

output "resource_group_name" {
  value = local.resource_group_name
}