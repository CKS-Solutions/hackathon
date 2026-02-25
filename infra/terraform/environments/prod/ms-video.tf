# ms-video: S3, SQS, DynamoDB for video processing

locals {
  ms_video_tags = {
    Project     = "hackathon-ms-video"
    Managed     = "terraform"
    Environment = var.environment
  }
}

# S3 bucket for video storage
resource "aws_s3_bucket" "video_system" {
  bucket = "video-system-${var.environment}"
  tags   = local.ms_video_tags
}

resource "aws_s3_bucket_versioning" "video_system" {
  bucket = aws_s3_bucket.video_system.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_lifecycle_configuration" "video_system" {
  bucket = aws_s3_bucket.video_system.id

  rule {
    id     = "delete-old-raw-videos"
    status = "Enabled"

    filter {
      prefix = "raw/"
    }

    expiration {
      days = 30
    }
  }

  rule {
    id     = "transition-processed-videos"
    status = "Enabled"

    filter {
      prefix = "processed/"
    }

    transition {
      days          = 90
      storage_class = "GLACIER"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "video_system" {
  bucket = aws_s3_bucket.video_system.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_cors_configuration" "video_system" {
  bucket = aws_s3_bucket.video_system.id

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["GET", "PUT", "POST"]
    allowed_origins = ["*"]
    expose_headers  = ["ETag"]
    max_age_seconds = 3000
  }
}

# SQS queue for video processing
module "ms_video_sqs" {
  source = "../../modules/sqs-queue"

  queue_name                 = "MSVideo-Queue"
  create_dlq                 = true
  dlq_name                   = "MSVideo-DLQueue"
  visibility_timeout_seconds = 300
  message_retention_seconds  = 345600
  receive_wait_time_seconds  = 20
  max_receive_count          = 3
  tags                       = local.ms_video_tags
}

# DynamoDB table for video metadata
resource "aws_dynamodb_table" "ms_video_videos" {
  name         = "MSVideo.Video"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "id"

  attribute {
    name = "id"
    type = "S"
  }

  attribute {
    name = "user_id"
    type = "S"
  }

  attribute {
    name = "created_at"
    type = "S"
  }

  global_secondary_index {
    name            = "user_id-index"
    hash_key        = "user_id"
    range_key       = "created_at"
    projection_type = "ALL"
  }

  tags = local.ms_video_tags
}

# IAM policy for ms-video service
resource "aws_iam_policy" "ms_video_policy" {
  name        = "MSVideoServicePolicy-${var.environment}"
  description = "IAM policy for ms-video service"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:DeleteObject",
          "s3:ListBucket"
        ]
        Resource = [
          aws_s3_bucket.video_system.arn,
          "${aws_s3_bucket.video_system.arn}/*"
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "sqs:SendMessage",
          "sqs:ReceiveMessage",
          "sqs:DeleteMessage",
          "sqs:GetQueueAttributes"
        ]
        Resource = [
          module.ms_video_sqs.queue_arn,
          module.ms_video_sqs.dlq_arn
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "dynamodb:GetItem",
          "dynamodb:PutItem",
          "dynamodb:UpdateItem",
          "dynamodb:Query",
          "dynamodb:Scan"
        ]
        Resource = [
          aws_dynamodb_table.ms_video_videos.arn,
          "${aws_dynamodb_table.ms_video_videos.arn}/index/*"
        ]
      }
    ]
  })

  tags = local.ms_video_tags
}

# Attach policy to the app service account role
resource "aws_iam_role_policy_attachment" "ms_video_policy_attachment" {
  role       = aws_iam_role.app_role.name
  policy_arn = aws_iam_policy.ms_video_policy.arn
}

# Outputs
output "ms_video_bucket_name" {
  description = "S3 bucket name for video storage"
  value       = aws_s3_bucket.video_system.id
}

output "ms_video_bucket_arn" {
  description = "S3 bucket ARN for video storage"
  value       = aws_s3_bucket.video_system.arn
}

output "ms_video_queue_url" {
  description = "SQS queue URL for video processing"
  value       = module.ms_video_sqs.queue_url
}

output "ms_video_queue_arn" {
  description = "SQS queue ARN for video processing"
  value       = module.ms_video_sqs.queue_arn
}

output "ms_video_table_name" {
  description = "DynamoDB table name for video metadata"
  value       = aws_dynamodb_table.ms_video_videos.name
}

output "ms_video_table_arn" {
  description = "DynamoDB table ARN for video metadata"
  value       = aws_dynamodb_table.ms_video_videos.arn
}
