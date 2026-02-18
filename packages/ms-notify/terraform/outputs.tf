output "sqs_queue_url" {
  value = aws_sqs_queue.main.id
}

output "sqs_queue_arn" {
  value = aws_sqs_queue.main.arn
}

output "dynamodb_table_name" {
  value = aws_dynamodb_table.main.name
}

output "ses_sender_email" {
  value = aws_ses_email_identity.sender.email
}
