terraform {
  required_providers {
    azurerm = {
      source = "opentofu/azurerm"
    }
  }
}

data "azurerm_client_config" "current" {}

resource "azurerm_resource_group" "shared" {
  count = var.create_resource_group ? 1 : 0

  name     = var.resource_group_name
  location = var.location
}

data "azurerm_resource_group" "shared" {
  count = var.create_resource_group ? 0 : 1

  name = var.resource_group_name
}

locals {
  resource_group_name = var.create_resource_group ? azurerm_resource_group.shared[0].name : data.azurerm_resource_group.shared[0].name
}

resource "azurerm_storage_account" "tfstate" {
  count = var.create_storage_account ? 1 : 0

  # Azure storage account names are max 24 chars: 17 (prefix) + 7 ("tfstate").
  # Keep in sync with internal/cloud/azure StateStorageAccountName.
  name                     = "${substr(replace(var.name_prefix, "-", ""), 0, 17)}tfstate"
  resource_group_name      = local.resource_group_name
  location                 = var.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
  min_tls_version          = "TLS1_2"
}

data "azurerm_storage_account" "tfstate" {
  count               = var.create_storage_account ? 0 : 1
  name                = "${substr(replace(var.name_prefix, "-", ""), 0, 17)}tfstate"
  resource_group_name = local.resource_group_name
}

locals {
  storage_account_id = var.create_storage_account ? azurerm_storage_account.tfstate[0].id : data.azurerm_storage_account.tfstate[0].id
}

resource "azurerm_storage_container" "tfstate" {
  count = var.create_container ? 1 : 0

  name                  = "tfstate"
  storage_account_id    = local.storage_account_id
  container_access_type = "private"
}

# OpenTofu azurerm remote state uses use_azuread_auth; Contributor on the sub is not enough.
resource "azurerm_role_assignment" "tfstate_blob_contributor" {
  scope                = local.storage_account_id
  role_definition_name = "Storage Blob Data Contributor"
  principal_id         = data.azurerm_client_config.current.object_id
}