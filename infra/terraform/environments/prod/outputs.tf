output "ecr_repository_urls" {
  description = "Map of ECR repository names to URLs. Use these for GitHub Actions variables (ECR_MS_AUTH_URL, etc.)."
  value       = module.ecr.repository_urls
}

output "github_actions_role_arn" {
  description = "ARN of the IAM role for GitHub Actions OIDC. Set as secret AWS_ROLE_ARN in the repo."
  value       = aws_iam_role.github_actions_ecr.arn
}

# output "vpc_id" {
#   description = "ID of the VPC."
#   value       = module.vpc.vpc_id
# }

# output "cluster_id" {
#   description = "EKS cluster ID (name)."
#   value       = module.eks.cluster_id
# }

# output "cluster_endpoint" {
#   description = "Endpoint for the EKS cluster API."
#   value       = module.eks.cluster_endpoint
# }

# output "kubeconfig_command" {
#   description = "Command to update kubeconfig for the EKS cluster."
#   value       = "aws eks update-kubeconfig --region ${var.aws_region} --name ${var.cluster_name}"
# }
