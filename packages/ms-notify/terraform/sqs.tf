# resource "aws_sqs_queue" "dlq" {
#   name = "MSNotify-DLQueue"

#   message_retention_seconds = 1209600

#   tags = local.tags
# }

resource "aws_sqs_queue" "main" {
  name = "MSNotify-Queue"

  visibility_timeout_seconds = 60
  message_retention_seconds = 345600
  receive_wait_time_seconds = 20

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.dlq.arn
    maxReceiveCount     = 5
  })

  tags = local.tags
}
