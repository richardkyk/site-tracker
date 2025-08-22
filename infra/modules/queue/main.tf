variable "app_name" {
  type = string
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

output "env" {
  value = {
    SQS_URL = aws_sqs_queue.email_queue.id
  }
}
