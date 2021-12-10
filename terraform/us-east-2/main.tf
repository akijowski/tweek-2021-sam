terraform {
  required_version = "~> 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 3.0"
    }
  }
  backend "s3" {
    region         = "us-east-2"
    dynamodb_table = "sg-akijowski-backend-terraform"
    bucket         = "sg-akijowski-backend-terraform"
    key            = "us-east-2/tweekweek/state.tfstate"
  }
}

provider "aws" {
  region = "us-east-2"
  default_tags {
    tags = {
      MadeBy  = "Adam_K"
      MadeFor = "Tweek-Week 2021"
    }
  }
}

module "dynamodb" {
  source            = "../modules/dynamodb"
  dynamo_table_name = var.dynamo_table_name
  dynamo_hash_key   = var.dynamo_hash_key
  dynamo_range_key  = var.dynamo_range_key
}

module "iam_role" {
  source                 = "../modules/iam"
  dynamo_table_name      = var.dynamo_table_name
  lambda_name            = var.lambda_name
  enable_basic_execution = true
  enable_dynamo_access   = true
}