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
