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
  image_tag = "v1"
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

}

resource "aws_apigatewayv2_api" "http_api" {
  name          = "${var.app_name}-http-api"
  protocol_type = "HTTP"
}

resource "aws_apigatewayv2_integration" "lambda_integration" {
  api_id                 = aws_apigatewayv2_api.http_api.id
  integration_type       = "AWS_PROXY"
  integration_uri        = aws_lambda_function.lambda.invoke_arn
  integration_method     = "POST"
  payload_format_version = "2.0"
}

resource "aws_lambda_permission" "api_gw_invoke" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.lambda.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.http_api.execution_arn}/*/*"
}

resource "aws_lambda_permission" "api_gw_invoke_custom_domain" {
  statement_id  = "AllowInvokeFromCustomDomain"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.lambda.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_domain_name.custom.arn}/*"
}

resource "aws_apigatewayv2_route" "default_route" {
  api_id    = aws_apigatewayv2_api.http_api.id
  route_key = "ANY /{proxy+}" # catch all
  target    = "integrations/${aws_apigatewayv2_integration.lambda_integration.id}"
}

resource "aws_apigatewayv2_stage" "default_stage" {
  api_id      = aws_apigatewayv2_api.http_api.id
  name        = "$default"
  auto_deploy = true

  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.api_logs.arn
    format = jsonencode({
      requestId         = "$context.requestId"
      requestTime       = "$context.requestTime"
      httpMethod        = "$context.httpMethod"
      path              = "$context.path"
      status            = "$context.status"
      integrationStatus = "$context.integrationStatus"
      responseLength    = "$context.responseLength"
      errorMessage      = "$context.error.message"
    })
  }
}

output "api_url" {
  value = aws_apigatewayv2_stage.default_stage.invoke_url
}

resource "aws_apigatewayv2_domain_name" "custom" {
  domain_name = var.api_domain_name
  domain_name_configuration {
    certificate_arn = aws_acm_certificate_validation.cert_validation.certificate_arn
    endpoint_type   = "REGIONAL"
    security_policy = "TLS_1_2"
  }
}

resource "aws_apigatewayv2_api_mapping" "custom_mapping" {
  api_id          = aws_apigatewayv2_api.http_api.id
  domain_name     = aws_apigatewayv2_domain_name.custom.domain_name
  stage           = aws_apigatewayv2_stage.default_stage.name
  api_mapping_key = var.api_mapping_key
}

resource "aws_acm_certificate" "cert" {
  domain_name       = var.api_domain_name
  validation_method = "DNS"
}

output "certificate_domain_validation_options" {
  value = aws_acm_certificate.cert.domain_validation_options
}

output "api_gateway_custom_domain_target" {
  value = aws_apigatewayv2_domain_name.custom.domain_name_configuration[0].target_domain_name
}

resource "aws_acm_certificate_validation" "cert_validation" {
  certificate_arn = aws_acm_certificate.cert.arn
}

resource "aws_cloudwatch_log_group" "api_logs" {
  name              = "/aws/apigateway/${var.app_name}-access"
  retention_in_days = 7
}
