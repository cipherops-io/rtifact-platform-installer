terraform {
  required_providers {
    azurerm = {
      source = "opentofu/azurerm"
    }
  }
}

data "azurerm_client_config" "current" {}

# Bootstrap or a prior apply may have created the RG while state still tracks the
# managed resource; dropping create_resource_group must not delete the live RG.
removed {
  from = azurerm_resource_group.shared

  lifecycle {
    destroy = false
  }
}

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
  # Azure storage account names are max 24 chars: 17 (prefix) + 7 ("tfstate").
  # Keep in sync with internal/cloud/azure StateStorageAccountName.
  name                     = "${substr(replace(var.name_prefix, "-", ""), 0, 17)}tfstate"
  resource_group_name      = local.resource_group_name
  location                 = var.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
  min_tls_version          = "TLS1_2"
}

resource "azurerm_storage_container" "tfstate" {
  name                  = "tfstate"
  storage_account_id    = azurerm_storage_account.tfstate.id
  container_access_type = "private"
}

# OpenTofu azurerm remote state uses use_azuread_auth; Contributor on the sub is not enough.
resource "azurerm_role_assignment" "tfstate_blob_contributor" {
  scope                = azurerm_storage_account.tfstate.id
  role_definition_name = "Storage Blob Data Contributor"
  principal_id         = data.azurerm_client_config.current.object_id
}