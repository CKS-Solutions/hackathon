output "queue_url" {
  description = "URL of the main SQS queue."
  value       = aws_sqs_queue.main.url
}

output "queue_arn" {
  description = "ARN of the main SQS queue."
  value       = aws_sqs_queue.main.arn
}

output "dlq_arn" {
  description = "ARN of the dead-letter queue (only set when create_dlq is true)."
  value       = var.create_dlq ? aws_sqs_queue.dlq[0].arn : null
}
