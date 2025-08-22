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
  environment_vars = merge(
    module.db.env,
    module.queue.env,
  )
}
