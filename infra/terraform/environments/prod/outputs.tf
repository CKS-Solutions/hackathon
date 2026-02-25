output "ecr_repository_urls" {
  description = "Map of ECR repository names to full URLs. For GitHub Actions, set ECR_REGISTRY to the registry prefix (e.g. strip /ms-auth from any URL)."
  value       = module.ecr.repository_urls
}

output "github_actions_role_arn" {
  description = "ARN of the IAM role for GitHub Actions (build-push to ECR). Set as secret AWS_ROLE_ARN in the repo."
  value       = aws_iam_role.github_actions_ecr.arn
}

output "lb_controller_role_arn" {
  description = "ARN of the IAM role for AWS Load Balancer Controller (IRSA). Use for Helm install: serviceAccount.annotations.eks.amazonaws.com/role-arn."
  value       = aws_iam_role.lb_controller.arn
}

output "vpc_id" {
  description = "ID of the VPC."
  value       = module.vpc.vpc_id
}

output "cluster_id" {
  description = "EKS cluster ID (name)."
  value       = module.eks.cluster_id
}

output "cluster_endpoint" {
  description = "Endpoint for the EKS cluster API."
  value       = module.eks.cluster_endpoint
}

output "kubeconfig_command" {
  description = "Command to update kubeconfig for the EKS cluster."
  value       = "aws eks update-kubeconfig --region ${var.aws_region} --name ${var.cluster_name}"
}

# ms-notify: use these for Deployment env (ConfigMap/Secret or External Secrets)
output "ms_notify_sqs_queue_url" {
  description = "SQS queue URL for ms-notify. Set as env in ms-notify Deployment."
  value       = module.ms_notify_sqs.queue_url
}

output "ms_notify_dynamodb_table_name" {
  description = "DynamoDB table name for ms-notify. Set as env in ms-notify Deployment."
  value       = module.ms_notify_dynamodb_table.table_name
}

output "ms_notify_ses_sender_email" {
  description = "SES verified sender email for ms-notify. Set as env in ms-notify Deployment."
  value       = module.ms_notify_ses.email
}

# ms-video: use these for Deployment env (ConfigMap/Secret or External Secrets)
output "ms_video_bucket_name" {
  description = "S3 bucket name for video storage. Set as env in ms-video Deployment."
  value       = aws_s3_bucket.video_system.id
}

output "ms_video_sqs_queue_url" {
  description = "SQS queue URL for ms-video processing. Set as env in ms-video Deployment."
  value       = module.ms_video_sqs.queue_url
}

output "ms_video_dynamodb_table_name" {
  description = "DynamoDB table name for ms-video metadata. Set as env in ms-video Deployment."
  value       = aws_dynamodb_table.ms_video_videos.name
}

output "app_irsa_role_arn" {
  description = "ARN of the IAM role for video-system apps (IRSA). ServiceAccount video-system/app; used by ms-auth, ms-video, ms-notify."
  value       = aws_iam_role.app.arn
}

output "ms_auth_db_endpoint" {
  description = "RDS endpoint for ms-auth database. Use for connection string."
  value       = module.ms_auth_rds.db_instance_endpoint
}

output "ms_auth_db_address" {
  description = "RDS hostname for ms-auth database."
  value       = module.ms_auth_rds.db_instance_address
}

output "ms_auth_db_port" {
  description = "RDS port for ms-auth database."
  value       = module.ms_auth_rds.db_instance_port
}

output "ms_auth_db_name" {
  description = "Database name for ms-auth."
  value       = module.ms_auth_rds.db_name
}

output "ms_auth_db_secret_arn" {
  description = "ARN of the Secrets Manager secret containing ms-auth database credentials. Use with External Secrets Operator or direct access from pods."
  value       = module.ms_auth_rds.secret_arn
}

output "ms_auth_db_secret_name" {
  description = "Name of the Secrets Manager secret containing ms-auth database credentials."
  value       = module.ms_auth_rds.secret_name
}

output "ms_auth_db_security_group_id" {
  description = "Security group ID for ms-auth RDS instance."
  value       = module.ms_auth_rds.security_group_id
}

# Shared secrets
output "jwt_secret_arn" {
  description = "ARN of the Secrets Manager secret containing JWT secret shared by ms-auth and ms-video."
  value       = aws_secretsmanager_secret.jwt_secret.arn
}

output "jwt_secret_name" {
  description = "Name of the Secrets Manager secret containing JWT secret."
  value       = aws_secretsmanager_secret.jwt_secret.name
}
