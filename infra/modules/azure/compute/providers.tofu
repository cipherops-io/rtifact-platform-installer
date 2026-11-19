terraform {
  required_version = ">= 1.6.0"
  required_providers {
    azurerm = {
      source = "opentofu/azurerm"
    }
    tls = {
      source  = "opentofu/tls"
      version = ">= 4.0"
    }
  }
}