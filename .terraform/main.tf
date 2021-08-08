terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 3.27"
    }
  }

  required_version = ">= 0.14.9"

  backend "remote" {
    organization = "onetwentyseven"

    workspaces {
      name = "AWS"
    }
  }
}

provider "aws" {
  profile = "default"
  region  = "us-east-1"
}

