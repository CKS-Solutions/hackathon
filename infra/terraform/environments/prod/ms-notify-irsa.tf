# IRSA for ms-notify: pod uses this role to call SQS, DynamoDB, SES. ServiceAccount video-system/ms-notify must have annotation eks.amazonaws.com/role-arn.

resource "aws_iam_role" "ms_notify" {
  name = "${var.cluster_name}-ms-notify"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Federated = module.eks.oidc_provider_arn
        }
        Action = "sts:AssumeRoleWithWebIdentity"
        Condition = {
          StringEquals = {
            "${local.oidc_id}:sub" = "system:serviceaccount:video-system:ms-notify"
          }
        }
      }
    ]
  })
}

data "aws_iam_policy_document" "ms_notify" {
  statement {
    sid    = "SQS"
    effect = "Allow"
    actions = [
      "sqs:ReceiveMessage",
      "sqs:SendMessage",
      "sqs:DeleteMessage",
      "sqs:GetQueueAttributes"
    ]
    resources = [module.ms_notify_sqs.queue_arn]
  }

  statement {
    sid    = "DynamoDB"
    effect = "Allow"
    actions = [
      "dynamodb:PutItem",
      "dynamodb:GetItem",
      "dynamodb:UpdateItem",
      "dynamodb:Query",
      "dynamodb:BatchGetItem",
      "dynamodb:BatchWriteItem"
    ]
    resources = [
      module.ms_notify_dynamodb_table.table_arn,
      "${module.ms_notify_dynamodb_table.table_arn}/*"
    ]
  }

  statement {
    sid    = "SES"
    effect = "Allow"
    actions = [
      "ses:SendEmail",
      "ses:SendRawEmail"
    ]
    resources = [module.ms_notify_ses.arn]
  }
}

resource "aws_iam_role_policy" "ms_notify" {
  name   = "${var.cluster_name}-ms-notify"
  role   = aws_iam_role.ms_notify.id
  policy = data.aws_iam_policy_document.ms_notify.json
}
