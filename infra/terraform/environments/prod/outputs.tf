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

output "app_irsa_role_arn" {
  description = "ARN of the IAM role for video-system apps (IRSA). ServiceAccount video-system/app; used by ms-auth, ms-video, ms-notify."
  value       = aws_iam_role.app.arn
}
