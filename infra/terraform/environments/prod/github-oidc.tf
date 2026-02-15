# OIDC + IAM for GitHub Actions (push to ECR without long-lived credentials)

data "aws_caller_identity" "current" {}

data "tls_certificate" "github" {
  url = "https://token.actions.githubusercontent.com"
}

resource "aws_iam_openid_connect_provider" "github" {
  url             = "https://token.actions.githubusercontent.com"
  client_id_list  = ["sts.amazonaws.com"]
  thumbprint_list = [data.tls_certificate.github.certificates[0].sha1_fingerprint]
}

data "aws_iam_policy_document" "github_oidc_assume" {
  statement {
    sid     = "AllowGitHubOIDC"
    effect  = "Allow"
    actions = ["sts:AssumeRoleWithWebIdentity"]

    principals {
      type        = "Federated"
      identifiers = [aws_iam_openid_connect_provider.github.arn]
    }

    condition {
      test     = "StringEquals"
      variable = "token.actions.githubusercontent.com:aud"
      values   = ["sts.amazonaws.com"]
    }

    condition {
      test     = "StringLike"
      variable = "token.actions.githubusercontent.com:sub"
      values   = ["repo:${var.github_repo}:*"]
    }
  }
}

resource "aws_iam_role" "github_actions_ecr" {
  name               = "github-actions-ecr"
  assume_role_policy = data.aws_iam_policy_document.github_oidc_assume.json
}

data "aws_iam_policy_document" "github_actions_ecr" {
  statement {
    sid    = "ECRGetAuthorizationToken"
    effect = "Allow"
    actions = [
      "ecr:GetAuthorizationToken"
    ]
    resources = ["*"]
  }

  statement {
    sid    = "ECRPushPull"
    effect = "Allow"
    actions = [
      "ecr:BatchGetImage",
      "ecr:BatchCheckLayerAvailability",
      "ecr:PutImage",
      "ecr:InitiateLayerUpload",
      "ecr:UploadLayerPart",
      "ecr:CompleteLayerUpload"
    ]
    resources = values(module.ecr.repository_arns)
  }
}

# S3 permissions for Terraform state (backend) — same role used by terraform-apply/destroy workflows
data "aws_iam_policy_document" "github_actions_tf_state" {
  statement {
    sid    = "TerraformStateBucketList"
    effect = "Allow"
    actions = [
      "s3:ListBucket"
    ]
    resources = ["arn:aws:s3:::${var.terraform_state_bucket}"]
  }

  statement {
    sid    = "TerraformStateObject"
    effect = "Allow"
    actions = [
      "s3:GetObject",
      "s3:HeadObject",
      "s3:PutObject",
      "s3:DeleteObject"
    ]
    resources = ["arn:aws:s3:::${var.terraform_state_bucket}/prod/*"]
  }
}

resource "aws_iam_role_policy" "github_actions_ecr" {
  name   = "ecr-push"
  role   = aws_iam_role.github_actions_ecr.id
  policy = data.aws_iam_policy_document.github_actions_ecr.json
}

resource "aws_iam_role_policy" "github_actions_tf_state" {
  name   = "tf-state-s3"
  role   = aws_iam_role.github_actions_ecr.id
  policy = data.aws_iam_policy_document.github_actions_tf_state.json
}
