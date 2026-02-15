output "repository_urls" {
  description = "Map of repository name to ECR repository URL."
  value       = { for k, r in aws_ecr_repository.this : k => r.repository_url }
}

output "repository_arns" {
  description = "Map of repository name to ECR repository ARN (for IAM policies)."
  value       = { for k, r in aws_ecr_repository.this : k => r.arn }
}
