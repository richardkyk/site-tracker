terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.7"
    }
  }
  required_version = ">= 1.12.0"
}

variable "custom_domain" {
  description = "Custom domain to use for SES"
  type        = string
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

module "storage" {
  source = "./modules/storage"
  name   = "${local.app_name}-storage"
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
  s3_bucket_arns = [module.storage.arn]
  environment_vars = merge(
    {
      SQS_TASK_URL   = module.task_queue.id
      DYNAMODB_TABLE = module.db.name
      S3_BUCKET      = module.storage.name
    }
  )
}

output "manager_api_url" {
  description = "API Gateway URL for Manager Lambda"
  value       = module.manager_lambda.api_url
}

module "cron_schedule" {
  source = "./modules/cron"

  name     = "${local.app_name}-schedule"
  schedule = "cron(0 */4 * * ? *)"
}

module "cron_lambda" {
  source = "./modules/lambda"

  function_name   = "${local.app_name}-cron"
  filename        = "../build/cron.zip"
  dynamodb_arns   = [module.db.arn]
  allowed_methods = ["GET"]
  sqs_target_arns = [module.task_queue.arn]
  cron_triggers = [
    {
      arn  = module.cron_schedule.arn
      name = module.cron_schedule.name
  }]
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
  s3_bucket_arns   = [module.storage.arn]
  environment_vars = merge(
    {
      SQS_NOTIFY_URL = module.notify_queue.id
      DYNAMODB_TABLE = module.db.name
      S3_BUCKET      = module.storage.name
    }
  )
}

data "aws_ses_domain_identity" "example" {
  domain = var.custom_domain
}

module "mailer_lambda" {
  source = "./modules/lambda"

  function_name    = "${local.app_name}-mailer"
  filename         = "../build/mailer.zip"
  sqs_trigger_arns = [module.notify_queue.arn]
  ses_arns         = [data.aws_ses_domain_identity.example.arn]
  environment_vars = merge(
    {
      SES_FROM_EMAIL = "no-reply@${data.aws_ses_domain_identity.example.domain}"
    }
  )
}
