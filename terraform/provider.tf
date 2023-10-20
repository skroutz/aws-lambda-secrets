terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.21.0"
    }
  }
}


provider "aws" {
  region = "eu-central-1"

  default_tags {
    tags = {
      DeployedFrom = "https://github.com/skroutz/aws-lambda-secrets"
      Team         = "Security"
      Environment  = "Staging"
      ManagedBy    = "Terraform" 
    }
  }
}
