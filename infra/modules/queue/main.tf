variable "app_name" {
  type = string
}

resource "aws_sqs_queue" "this" {
  name = "${var.app_name}-queue"
}

output "name" {
  value = aws_sqs_queue.this.name
}

output "arn" {
  value = aws_sqs_queue.this.arn
}

output "env" {
  value = {
    SQS_URL = aws_sqs_queue.this.id
  }
}
