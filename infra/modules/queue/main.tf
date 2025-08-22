variable "app_name" {
  type = string
}

resource "aws_sqs_queue" "dlq" {
  name = "${var.app_name}-dlq"
}

resource "aws_sqs_queue" "this" {
  name = "${var.app_name}-queue"

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.dlq.arn
    maxReceiveCount     = 3 # after 3 failed receives, message goes to DLQ
  })
}

output "name" {
  value = aws_sqs_queue.this.name
}

output "arn" {
  value = aws_sqs_queue.this.arn
}

output "id" {
  value = aws_sqs_queue.this.id
}

