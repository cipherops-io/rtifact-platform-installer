terraform {
  required_version = ">= 1.6.0"
  required_providers {
    azurerm = {
      source = "opentofu/azurerm"
    }
    random = {
      source  = "opentofu/random"
      version = ">= 3.0"
    }
  }
}