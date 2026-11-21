terraform {
  required_providers {
    aws = {
      source                = "opentofu/aws"
      configuration_aliases = [aws.us_east_1]
    }
  }
}