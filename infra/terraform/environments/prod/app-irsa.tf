# IRSA genérico para os 3 serviços (ms-auth, ms-video, ms-notify). ServiceAccount video-system/app com annotation eks.amazonaws.com/role-arn.

resource "aws_iam_role" "app" {
  name = "${var.cluster_name}-app"

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
            "${local.oidc_id}:sub" = "system:serviceaccount:video-system:app"
          }
        }
      }
    ]
  })
}

# Permissões amplas (hackathon) para evitar 403. Depois trocar por políticas mínimas por recurso.
data "aws_iam_policy_document" "app" {
  # SQS: todas as filas da conta na região
  statement {
    sid    = "SQS"
    effect = "Allow"
    actions = [
      "sqs:*"
    ]
    resources = [
      "arn:aws:sqs:${var.aws_region}:${data.aws_caller_identity.current.account_id}:*"
    ]
  }

  # DynamoDB: todas as tabelas da conta na região
  statement {
    sid    = "DynamoDB"
    effect = "Allow"
    actions = [
      "dynamodb:*"
    ]
    resources = [
      "arn:aws:dynamodb:${var.aws_region}:${data.aws_caller_identity.current.account_id}:table/*",
      "arn:aws:dynamodb:${var.aws_region}:${data.aws_caller_identity.current.account_id}:table/*/index/*"
    ]
  }

  # SES: envio de e-mail (recurso * é o usual para SendEmail/SendRawEmail)
  statement {
    sid    = "SES"
    effect = "Allow"
    actions = [
      "ses:SendEmail",
      "ses:SendRawEmail",
      "ses:GetSendQuota",
      "ses:GetSendStatistics"
    ]
    resources = ["*"]
  }
}

resource "aws_iam_role_policy" "app" {
  name   = "${var.cluster_name}-app"
  role   = aws_iam_role.app.id
  policy = data.aws_iam_policy_document.app.json
}
