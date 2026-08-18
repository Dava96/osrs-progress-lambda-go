terraform {
  backend "s3" {
    bucket       = "osrs-progress-tfstate-491234348360"
    key          = "osrs-progress-lambda/terraform.tfstate"
    region       = "eu-west-1"
    encrypt      = true
    use_lockfile = true
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "6.36.0"
    }
  }

  required_version = ">= 1.15"
}
