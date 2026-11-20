# GENERATED from vars/versions.yml — edit there, then: make bump-providers
terraform {
  required_version = ">= 1.13.9"
  required_providers {
    google = {
      source  = "opentofu/google"
      version = "8.12.0"
    }
    random = {
      source  = "opentofu/random"
      version = "3.9.1"
    }
    tls = {
      source  = "opentofu/tls"
      version = "4.4.2"
    }
  }
}
