provider "azurerm" {
  features {
    key_vault {
      purge_soft_delete_on_destroy = false
    }
    resource_group {
      prevent_deletion_if_contains_resources = false
    }
  }

  subscription_id                 = var.subscription_id
  environment                     = var.azure_environment
  use_msi                         = var.use_msi
  use_cli                         = !var.use_msi
  resource_provider_registrations = "legacy"
}
