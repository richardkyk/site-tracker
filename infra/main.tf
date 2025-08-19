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
  image_tag = "v4"
}

# Create ECR repository to hold your Docker image
resource "aws_ecr_repository" "lambda_repo" {
  name = "${var.app_name}-repo"
}

# Build and push Docker image to ECR
resource "null_resource" "docker_build_and_push" {
  depends_on = [aws_ecr_repository.lambda_repo]

  triggers = {
    image_tag = local.image_tag
  }

  provisioner "local-exec" {
    command = <<EOT
      aws ecr get-login-password --region ${var.aws_region} | docker login --username AWS --password-stdin ${aws_ecr_repository.lambda_repo.repository_url}
      docker build -f ../Dockerfile -t ${aws_ecr_repository.lambda_repo.repository_url}:${local.image_tag} ../
      docker push ${aws_ecr_repository.lambda_repo.repository_url}:${local.image_tag}
    EOT
  }
}

# Create IAM role for Lambda execution
resource "aws_iam_role" "lambda_exec" {
  name               = "${var.app_name}-lambda-exec-role"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json
}

data "aws_iam_policy_document" "lambda_assume_role" {
  statement {
    effect = "Allow"
    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
    actions = ["sts:AssumeRole"]
  }
}

# Attach managed policies to the role
resource "aws_iam_role_policy_attachment" "lambda_basic" {
  role       = aws_iam_role.lambda_exec.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

# Create Lambda function from container image
resource "aws_lambda_function" "lambda" {
  depends_on = [null_resource.docker_build_and_push]

  function_name = "${var.app_name}-lambda"
  package_type  = "Image"
  image_uri     = "${aws_ecr_repository.lambda_repo.repository_url}:${local.image_tag}"
  role          = aws_iam_role.lambda_exec.arn
  timeout       = 10
  memory_size   = 512

  environment {
    variables = {
      DYNAMODB_TABLE = aws_dynamodb_table.dynamodb_table.name
      SQS_URL        = aws_sqs_queue.email_queue.id
    }
  }
}

resource "aws_dynamodb_table" "dynamodb_table" {
  name         = "${var.app_name}-table"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "id"

  attribute {
    name = "id"
    type = "S"
  }
}

resource "aws_iam_role_policy" "lambda_dynamodb" {
  role = aws_iam_role.lambda_exec.id
  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Action = [
          "dynamodb:GetItem",
          "dynamodb:Query",
          "dynamodb:Scan",
          "dynamodb:PutItem",
          "dynamodb:UpdateItem"
        ],
        Effect   = "Allow",
        Resource = aws_dynamodb_table.dynamodb_table.arn
      },
    ]
  })
}

resource "aws_api_gateway_rest_api" "api" {
  name        = "${var.app_name}-rest-api"
  description = "REST API for Lambda proxy integration"
}

resource "aws_api_gateway_method" "method" {
  rest_api_id   = aws_api_gateway_rest_api.api.id
  resource_id   = aws_api_gateway_rest_api.api.root_resource_id
  http_method   = "ANY"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "lambda_proxy" {
  rest_api_id             = aws_api_gateway_rest_api.api.id
  resource_id             = aws_api_gateway_rest_api.api.root_resource_id
  http_method             = aws_api_gateway_method.method.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = aws_lambda_function.lambda.invoke_arn
}

resource "aws_lambda_permission" "apigw" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.lambda.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_api_gateway_rest_api.api.execution_arn}/*/*"
}

resource "aws_api_gateway_stage" "prod" {
  rest_api_id   = aws_api_gateway_rest_api.api.id
  stage_name    = "prod"
  deployment_id = aws_api_gateway_deployment.deployment.id
}

resource "aws_api_gateway_deployment" "deployment" {
  depends_on  = [aws_api_gateway_integration.lambda_proxy]
  rest_api_id = aws_api_gateway_rest_api.api.id
}

output "api_url" {
  value = "https://${aws_api_gateway_rest_api.api.id}.execute-api.${var.aws_region}.amazonaws.com/prod"
}

resource "aws_sqs_queue" "email_queue" {
  name = "${var.app_name}-email-queue"
}

resource "aws_iam_role_policy" "lambda_sqs" {
  role = aws_iam_role.lambda_exec.id
  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect   = "Allow",
        Action   = ["sqs:SendMessage"],
        Resource = aws_sqs_queue.email_queue.arn
      }
    ]
  })
}

output "sqs_url" {
  value = aws_sqs_queue.email_queue.id
}
