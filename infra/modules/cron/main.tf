variable "name" {
  type = string
}

variable "schedule" {
  type = string
}

resource "aws_cloudwatch_event_rule" "cron_rule" {
  name                = var.name
  schedule_expression = var.schedule
}

output "name" {
  value = aws_cloudwatch_event_rule.cron_rule.name
}

output "arn" {
  value = aws_cloudwatch_event_rule.cron_rule.arn
}
