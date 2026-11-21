provider "aws" {
  region = var.location
}

provider "aws" {
  alias  = "us_east_1"
  region = "us-east-1"
}
