terraform {
  required_providers {
    aws = {
      source = "opentofu/aws"
    }
    tls = {
      source  = "opentofu/tls"
      version = ">= 4.0"
    }
  }
}
