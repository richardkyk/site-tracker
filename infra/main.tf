terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.7"
    }
  }
  required_version = ">= 1.12.0"
}

locals {
  app_name   = "site-tracker"
  aws_region = "ap-southeast-2"
}

provider "aws" {
  region = local.aws_region
}

module "db" {
  source   = "./modules/db"
  app_name = local.app_name
}

module "queue" {
  source   = "./modules/queue"
  app_name = local.app_name
}

module "scraper_lambda" {
  source = "./modules/lambda"

  function_name = "${local.app_name}-scraper"
  filename      = "../build/scraper.zip"
  dynamodb_arns = [module.db.arn]
  sqs_arns      = [module.queue.arn]
  allowed_methods = [
    "GET",
    "POST",
    "DELETE",
    "OPTIONS",
  ]
  environment_vars = merge(
    module.db.env,
    module.queue.env,
  )
}

output "scraper_api_url" {
  description = "API Gateway URL for Scraper Lambda"
  value       = module.scraper_lambda.api_url
}
