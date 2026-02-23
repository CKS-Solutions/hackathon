# ms-notify: SQS, DynamoDB, SES. Reuses generic modules.

locals {
  ms_notify_tags = {
    Project     = "hackathon-ms-notify"
    Managed     = "terraform"
    Environment = var.environment
  }
}

module "ms_notify_sqs" {
  source = "../../modules/sqs-queue"

  queue_name                 = "MSNotify-Queue"
  create_dlq                 = true
  dlq_name                   = "MSNotify-DLQueue"
  visibility_timeout_seconds = 60
  message_retention_seconds  = 345600
  receive_wait_time_seconds  = 20
  max_receive_count          = 5
  tags                       = local.ms_notify_tags
}

module "ms_notify_dynamodb_table" {
  source = "../../modules/dynamodb-table"

  table_name    = "MSNotify.Notification"
  hash_key      = "id"
  hash_key_type = "S"
  billing_mode  = "PAY_PER_REQUEST"
  tags          = local.ms_notify_tags
}

module "ms_notify_ses" {
  source = "../../modules/ses-email-identity"

  email = "cks.hackathon.noreply@gmail.com"
}
