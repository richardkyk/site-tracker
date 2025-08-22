variable "function_name" {
  type = string
}

variable "filename" {
  type = string
}

variable "handler" {
  type    = string
  default = "bootstrap" # for Go custom runtime
}

variable "runtime" {
  type    = string
  default = "provided.al2"
}

variable "timeout" {
  type    = number
  default = 30
}

variable "memory_size" {
  type    = number
  default = 512
}

variable "environment_vars" {
  type    = map(string)
  default = {}
}

variable "dynamodb_arns" {
  type    = list(string)
  default = [] # <- default empty list
}

variable "sqs_target_arns" {
  type        = list(string)
  description = "List of SQS queue ARNs to allow sending messages to"
  default     = [] # <- default empty list
}

variable "sqs_trigger_arns" {
  type        = list(string)
  description = "List of SQS queue ARNs to allow triggering the Lambda"
  default     = [] # <- default empty list
}

variable "allowed_methods" {
  type    = list(string)
  default = []
}

# IAM role for Lambda
resource "aws_iam_role" "this" {
  name = "${var.function_name}-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      }
    ]
  })
}

# Basic execution policy
resource "aws_iam_role_policy_attachment" "basic_execution" {
  role       = aws_iam_role.this.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

# Lambda itself
resource "aws_lambda_function" "this" {
  function_name = var.function_name
  role          = aws_iam_role.this.arn
  handler       = var.handler
  runtime       = var.runtime
  timeout       = var.timeout
  memory_size   = var.memory_size

  filename         = var.filename
  source_code_hash = filebase64sha256(var.filename)

  environment {
    variables = var.environment_vars
  }
}

# CloudWatch log group
resource "aws_cloudwatch_log_group" "this" {
  name              = "/aws/lambda/${aws_lambda_function.this.function_name}"
  retention_in_days = 14
}

output "lambda_arn" {
  value = aws_lambda_function.this.arn
}

resource "aws_iam_role_policy" "dynamodb_access" {
  count = length(var.dynamodb_arns) # 0 if none passed
  role  = aws_iam_role.this.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "dynamodb:GetItem",
          "dynamodb:PutItem",
          "dynamodb:UpdateItem",
          "dynamodb:DeleteItem",
          "dynamodb:Scan",
          "dynamodb:Query"
        ]
        Resource = var.dynamodb_arns[count.index]
      }
    ]
  })
}

resource "aws_iam_role_policy" "lambda_sqs_target" {
  count = length(var.sqs_target_arns) # 0 if none passed
  role  = aws_iam_role.this.id

  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect   = "Allow",
        Action   = ["sqs:SendMessage"],
        Resource = var.sqs_target_arns[count.index]
      }
    ]
  })
}

module "lambda_apigw" {
  source            = "../apigw"
  count             = length(var.allowed_methods) > 0 ? 1 : 0
  allowed_methods   = var.allowed_methods
  function_name     = var.function_name
  lambda_invoke_arn = aws_lambda_function.this.invoke_arn
}

output "api_url" {
  value = length(module.lambda_apigw) > 0 ? module.lambda_apigw[0].api_url : null
}

resource "aws_lambda_event_source_mapping" "from_sqs" {
  count            = length(var.sqs_trigger_arns) # 0 if none passed
  event_source_arn = var.sqs_trigger_arns[count.index]
  function_name    = aws_lambda_function.this.arn # Lambda ARN
  batch_size       = 1                            # Number of messages per invocation
  enabled          = true
}

resource "aws_iam_role_policy" "lambda_sqs_event_source" {
  count = length(var.sqs_trigger_arns) # 0 if none passed
  role  = aws_iam_role.this.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "sqs:ReceiveMessage",
          "sqs:DeleteMessage",
          "sqs:GetQueueAttributes",
          "sqs:GetQueueUrl"
        ]
        Resource = var.sqs_trigger_arns[count.index]
      }
    ]
  })
}
