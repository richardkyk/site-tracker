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

module "task_queue" {
  source   = "./modules/queue"
  app_name = "${local.app_name}-task"
}

module "notify_queue" {
  source   = "./modules/queue"
  app_name = "${local.app_name}-notify"
}

module "manager_lambda" {
  source = "./modules/lambda"

  function_name   = "${local.app_name}-manager"
  filename        = "../build/manager.zip"
  dynamodb_arns   = [module.db.arn]
  sqs_target_arns = [module.task_queue.arn]
  allowed_methods = [
    "GET",
    "POST",
    "DELETE",
    "OPTIONS",
  ]
  environment_vars = merge(
    {
      SQS_TASK_URL   = module.task_queue.id
      DYNAMODB_TABLE = module.db.name
    }
  )
}

output "manager_api_url" {
  description = "API Gateway URL for Manager Lambda"
  value       = module.manager_lambda.api_url
}

module "cron_lambda" {
  source = "./modules/lambda"

  function_name   = "${local.app_name}-cron"
  filename        = "../build/cron.zip"
  dynamodb_arns   = [module.db.arn]
  allowed_methods = ["GET"]
  sqs_target_arns = [module.task_queue.arn]
  environment_vars = merge(
    {
      SQS_TASK_URL   = module.task_queue.id
      DYNAMODB_TABLE = module.db.name
    }
  )
}

output "cron_api_url" {
  description = "API Gateway URL for Cron Lambda"
  value       = module.cron_lambda.api_url
}

module "scraper_lambda" {
  source = "./modules/lambda"

  function_name    = "${local.app_name}-scraper"
  filename         = "../build/scraper.zip"
  dynamodb_arns    = [module.db.arn]
  sqs_target_arns  = [module.notify_queue.arn]
  sqs_trigger_arns = [module.task_queue.arn]
  environment_vars = merge(
    {
      SQS_NOTIFY_URL = module.notify_queue.id
      DYNAMODB_TABLE = module.db.name
    }
  )
}
